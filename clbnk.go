// Clbnk implements export/import operations with client bank.
package clbnk

import (
	"fmt"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	HEADER       = "1CClientBankExchange"
	EXCH_VERSION = "1.03"
	DEF_SENDER   = "Бухгалтерия предприятия, редакция 3.0"
	FOOTER       = "КонецФайла"

	ENCODING_WIN = "Windows"
	ENCODING_DOS = "DOS"

	DOCUMENT_TYPE_ID = "Банковский ордер"
)
const (
	ENCODING_TYPE_WIN EncodingType = iota
	ENCODING_TYPE_DOS
	ENCODING_TYPE_NOT_DEFINED
)

type EncodingType int

func (e EncodingType) Marshal() ([]byte, error) {
	v := [][]byte{[]byte(ENCODING_WIN), []byte(ENCODING_DOS)}
	return v[int(e)], nil
}

func (e *EncodingType) Unmarshal(data string) error {
	if data == ENCODING_WIN {
		*e = ENCODING_TYPE_WIN

	} else if data == ENCODING_DOS {
		*e = ENCODING_TYPE_DOS

	} else {
		return fmt.Errorf("encoding not defined")
	}
	return nil
}

func (e EncodingType) decode(s []byte) ([]byte, error) {
	var char_map *charmap.Charmap
	if e == ENCODING_TYPE_WIN {
		char_map = charmap.Windows1251
	} else {
		char_map = charmap.CodePage866
	}
	dec := char_map.NewDecoder()
	out, err := dec.Bytes(s)
	if err != nil {
		return []byte{}, err
	}
	return out, nil
}

func (e EncodingType) encode(s []byte) ([]byte, error) {
	var char_map *charmap.Charmap
	if e == ENCODING_TYPE_WIN {
		char_map = charmap.Windows1251
	} else {
		char_map = charmap.CodePage866
	}
	enc := char_map.NewEncoder()
	return enc.Bytes(s)
}

// DocumentType
type DocumentType int

func DocumentTypeValues() []string {
	return []string{"Платежное поручение",
		"Банковский ордер",
	}
}

func (d DocumentType) Marshal() ([]byte, error) {
	return []byte(DocumentTypeValues()[int(d)]), nil
}

const (
	DOCUMENT_TYPE_PP DocumentType = iota
	DOCUMENT_TYPE_BANK_ORDER
)

// PayType
type PayType int

func (d PayType) Marshal() ([]byte, error) {
	v := [][]byte{[]byte("Электронно")}
	return v[int(d)], nil
}

const (
	PAY_TYPE_DIG PayType = iota
)

type BankImportDocument interface {
	GetType() DocumentType
}

type BankExportDocument interface {
	GetType() DocumentType
	GetDate() time.Time
}

type Account struct {
	DateFrom     time.Time `bank:"ДатаНачала"`
	DateTo       time.Time `bank:"ДатаКонца"`
	Account      string    `bank:"РасчСчет"`
	BalanceStart float64   `bank:"НачальныйОстаток"`
	BalanceEnd   float64   `bank:"КонечныйОстаток"`
	Debet        float64   `bank:"ВсегоПоступило"`
	Kredit       float64   `bank:"ВсегоСписано"`
}

// BankExport is the main structure for exporting bank documents.
type BankImport struct {
	Version      string               `bank:"ВерсияФормата"`
	EncodingType EncodingType         `bank:"Кодировка"`
	Sender       string               `bank:"Отправитель"`
	CreateDate   time.Time            `bank:"ДатаСоздания"`
	CreateTime   string               `bank:"ВремяСоздания"`
	DateFrom     time.Time            `bank:"ДатаНачала"`
	DateTo       time.Time            `bank:"ДатаКонца"`
	Account      string               `bank:"РасчСчет"`
	AccSection   []Account            `bankElemStart:"СекцияРасчСчет" bankElemEnd:"КонецРасчСчет"`
	Documents    []BankImportDocument `bankElemStart:"СекцияДокумент" bankElemEnd:"КонецДокумента"`
}

func NewBankImport() *BankImport {
	return &BankImport{}
}

// BankExport is the main structure for exporting bank documents.
type BankExport struct {
	Version       string               `bank:"ВерсияФормата"`
	EncodingType  EncodingType         `bank:"Кодировка"`
	Sender        string               `bank:"Отправитель"`
	CreateDate    time.Time            `bank:"ДатаСоздания"`
	CreateTime    string               `bank:"ВремяСоздания"`
	DateFrom      time.Time            `bank:"ДатаНачала"`
	DateTo        time.Time            `bank:"ДатаКонца"`
	DocumentTypes []DocumentType       `bankElemStart:"Документ=" bankElemEnd:"\r\n"`
	Documents     []BankExportDocument `bankElemStart:"СекцияДокумент\r\n" bankElemEnd:"КонецДокумента\r\n"`
}

func NewBankExport(documents []BankExportDocument) *BankExport {
	exp_data := &BankExport{Version: EXCH_VERSION,
		EncodingType: ENCODING_TYPE_WIN,
		CreateDate:   time.Now(),
		CreateTime:   time.Now().Format("15:04:05"),
		Sender:       DEF_SENDER,
		Documents:    documents,
	}
	return exp_data
}

// beforeMarshal adds some values to structure: DocumentTypes, DateFrom, DateTo
func (e *BankExport) beforeMarshal() {
	doc_uniq_types := make(map[DocumentType]struct{})
	for _, doc := range e.Documents {
		tp := doc.GetType()
		if _, ok := doc_uniq_types[tp]; !ok {
			e.DocumentTypes = append(e.DocumentTypes, tp)
			doc_uniq_types[tp] = struct{}{}
		}
		doc_date := doc.GetDate()
		if e.DateFrom.IsZero() || e.DateFrom.After(doc_date) {
			e.DateFrom = doc_date
		}
		if e.DateTo.IsZero() || e.DateTo.Before(doc_date) {
			e.DateTo = doc_date
		}
	}
}

// Marshal exports all documents.
func (e *BankExport) Marshal() ([]byte, error) {
	if len(e.Documents) == 0 {
		return nil, fmt.Errorf("no documents")
	}
	e.beforeMarshal()

	cont, err := marshal(e, "", "")
	if err != nil {
		return []byte{}, err
	}
	cont, err = e.EncodingType.encode(cont)
	if err != nil {
		return []byte{}, err
	}
	b := []byte(HEADER + "\n")
	b = append(b, cont...)
	b = append(b, []byte(FOOTER+"\n")...)
	return b, nil
}

// PPDocument is an export document structure for DOCUMENT_TYPE_PP.
type PPDocument struct {
	Num  int       `bank:"Номер"`
	Date time.Time `bank:"Дата"`
	Sum  float64   `bank:"Сумма"`
	// Payer      Payer
	// Receiver   Receiver
	Payer            string `bank:"Плательщик"`
	PayerInn         string `bank:"ПлательщикИНН"`
	PayerName        string `bank:"Плательщик1"`
	Payer2           string `bank:"Плательщик2"`
	Payer3           string `bank:"Плательщик3"`
	Payer4           string `bank:"Плательщик4"`
	PayerAccount     string `bank:"ПлательщикРасчСчет"`
	PayerBankName    string `bank:"ПлательщикБанк1"`
	PayerBankPlace   string `bank:"ПлательщикБанк2"`
	PayerBankBik     string `bank:"ПлательщикБИК"`
	PayerBankAccount string `bank:"ПлательщикКорсчет"`

	Receiver            string  `bank:"Получатель"`
	ReceiverInn         string  `bank:"ПолучательИНН"`
	ReceiverName        string  `bank:"Получатель1"`
	Receiver2           string  `bank:"Получатель2"`
	Receiver3           string  `bank:"Получатель3"`
	Receiver4           string  `bank:"Получатель4"`
	ReceiverAccount     string  `bank:"ПолучательСчет"`
	ReceiverBankName    string  `bank:"ПолучательБанк1"`
	ReceiverBankPlace   string  `bank:"ПолучательБанк2"`
	ReceiverBankBik     string  `bank:"ПолучательБИК"`
	ReceiverBankAccount string  `bank:"ПолучательКорсчет"`
	PayType             PayType `bank:"ВидПлатежа"` //вид платежа
	OplType             string  `bank:"ВидОплаты"`  //вид оплаты
	Order               int     `bank:"Очередность"`
	PayComment          string  `bank:"НазначениеПлатежа" lines:"6"`
}

func (d *PPDocument) GetType() DocumentType {
	return DOCUMENT_TYPE_PP
}

func (d *PPDocument) GetDate() time.Time {
	return d.Date
}

// BankOrderDocument is an import document structure for DOCUMENT_TYPE_BANK_ORDER.
type BankOrderDocument struct {
	Num           int       `bank:"Номер"`
	Date          time.Time `bank:"Дата"`
	Sum           float64   `bank:"Сумма"`
	ReceitDate    time.Time `bank:"КвитанцияДата"`
	ReceitTime    string    `bank:"КвитанцияВремя"`
	ReceitComment string    `bank:"КвитанцияСодержание"` // combined value

	Payer            string `bank:"Плательщик"`
	PayerInn         string `bank:"ПлательщикИНН"`
	PayerName        string `bank:"Плательщик1"`
	Payer2           string `bank:"Плательщик2"`
	Payer3           string `bank:"Плательщик3"`
	Payer4           string `bank:"Плательщик4"`
	PayerAccount     string `bank:"ПлательщикРасчСчет"`
	PayerBankName    string `bank:"ПлательщикБанк1"`
	PayerBankPlace   string `bank:"ПлательщикБанк2"`
	PayerBankBik     string `bank:"ПлательщикБИК"`
	PayerBankAccount string `bank:"ПлательщикКорсчет"`

	Receiver            string `bank:"Получатель"`
	ReceiverInn         string `bank:"ПолучательИНН"`
	Receiver2           string `bank:"Получатель2"`
	Receiver3           string `bank:"Получатель3"`
	Receiver4           string `bank:"Получатель4"`
	ReceiverAccount     string `bank:"ПолучательСчет"`
	ReceiverBankName    string `bank:"ПолучательБанк1"`
	ReceiverBankPlace   string `bank:"ПолучательБанк2"`
	ReceiverBankBik     string `bank:"ПолучательБИК"`
	ReceiverBankAccount string `bank:"ПолучательКорсчет"`

	KreditDate    time.Time `bank:"ДатаСписано"`
	DebetDate     time.Time `bank:"ДатаПоступило"`
	PayType       PayType   `bank:"ВидПлатежа"` //вид платежа
	Code          string    `bank:"Код"`
	PayDirectCode string    `bank:"КодНазПлатежа"`

	PayComment string `bank:"НазначениеПлатежа"`

	KBKValue         string `bank:"ПоказательКБК"`
	OKATOValue       string `bank:"ОКАТО"`
	OsnovanieValue   string `bank:"ПоказательОснования"`
	PeriodValue      string `bank:"ПоказательПериода"`
	NomerValue       string `bank:"ПоказательНомера"`
	DateValue        string `bank:"ПоказательДаты"`
	TipValue         string `bank:"ПоказательТипа"`
	Order            int    `bank:"Очередность"`
	AcceptTerm       string `bank:"СрокАкцепта"`
	AccredType       string `bank:"ВидАккредитива"`
	PayTerm          string `bank:"СрокПлатежа"`
	PayCond1         string `bank:"УсловиеОплаты1"`
	PayCond2         string `bank:"УсловиеОплаты2"`
	PayCond3         string `bank:"УсловиеОплаты3"`
	SupplierOrderNum string `bank:"НомерСчетаПоставщика"`
}

func (d *BankOrderDocument) GetType() DocumentType {
	return DOCUMENT_TYPE_BANK_ORDER
}
