package clbnk

import (
	"os"
	"testing"
	"time"
)

const (
	TEST_DOC_COUNT = 3

	TEST_DOC2_PAYER_ACC  = "40702810000000074935"
	TEST_DOC2_PAYER_INN  = "7123456777"
	TEST_DOC2_PAYER_NAME = `ООО "Пупкин и К"`
	TEST_DOC2_REC_INN    = "7123456789012"
	TEST_DOC2_REC_ACC    = "40702810267020000630"
	TEST_DOC2_SUM        = 150000.00

	TEST_DOC1_PAYER_ACC  = "40702810000000077777"
	TEST_DOC1_PAYER_INN  = "7123456789"
	TEST_DOC1_PAYER_NAME = `ООО "Рога и копыта"`
	TEST_DOC1_REC_INN    = "7123456789012"
	TEST_DOC1_REC_ACC    = "40702810267020000630"
	TEST_DOC1_SUM        = 13056.00

	TEST_DOC0_PAYER_ACC  = "40702810000000074935"
	TEST_DOC0_PAYER_INN  = "7777777777"
	TEST_DOC0_PAYER_NAME = `ООО "Наеб"`
	TEST_DOC0_REC_INN    = "7702070139"
	TEST_DOC0_REC_ACC    = "47422810119484000074"
	TEST_DOC0_SUM        = 6936.0
)

func TestExport(t *testing.T) {
	documents := []BankExportDocument{&PPDocument{Num: 1,
		Date:              time.Now(),
		Sum:               175000,
		PayerName:         `ООО "Рога и Копыта"`,
		PayerInn:          "1234567891",
		PayerAccount:      "12345678901234567890",
		PayerBankName:     "Объёббанк ОАО",
		PayerBankPlace:    "г. Москва",
		PayerBankBik:      "123456789",
		PayerBankAccount:  "12345678901234567890",
		ReceiverName:      `ИП Иванов А.А.`,
		ReceiverInn:       "111122223344",
		ReceiverAccount:   "12345678901234567890",
		ReceiverBankName:  "КакойтоБанк ОАО",
		ReceiverBankPlace: "г. Москва", ReceiverBankBik: "123456789", ReceiverBankAccount: "12345678901234567890", PayType: PAY_TYPE_DIG, OplType: "01",
		Order:      5,
		PayComment: "За товары, по счету №125 на сумму 175000-00",
	},
		&PPDocument{Num: 2,
			Date:             time.Now(),
			Sum:              375.25,
			PayerName:        `ООО "Рога и Копыта"`,
			PayerInn:         "1234567891",
			PayerAccount:     "12345678901234567890",
			PayerBankName:    "КакойтоБанк ОАО",
			PayerBankPlace:   "г. Москва",
			PayerBankBik:     "123456789",
			PayerBankAccount: "12345678901234567890",
			ReceiverName:     `ИП Иванов А.А.`,
			ReceiverInn:      "111122223344",
			ReceiverAccount:  "12345678901234567890",
			PayType:          PAY_TYPE_DIG,
			OplType:          "01",
			Order:            5,
			PayComment:       "За товары, по счету №777 на сумму 375-25\nPlus NDS 111-16",
		},
	}

	fl := NewBankExport(documents)
	fl.EncodingType = ENCODING_TYPE_WIN
	f, err := os.Create("to_bank.txt")
	if err != nil {
		t.Fatalf("os.Create() failed: %v", err)
	}
	b, err := fl.Marshal()
	if err != nil {
		t.Fatalf("ExchFile.String() failed: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("No data")
	}
	if _, err := f.Write(b); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}

func TestImport(t *testing.T) {
	f_cont, err := os.ReadFile("kl_to_1c.txt")
	if err != nil {
		panic(err)
	}

	imp := NewBankImport()
	imp.EncodingType = ENCODING_TYPE_WIN
	if err := imp.Unmarshal(f_cont); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(imp.Documents) != TEST_DOC_COUNT {
		t.Fatalf("document count failed, expected %d, got %d", TEST_DOC_COUNT, len(imp.Documents))
	}
	doc, ok := imp.Documents[0].(*BankOrderDocument)
	if !ok {
		t.Fatal("document[0] must be of type BankOrderDocument")
	}

	if doc.PayerAccount != TEST_DOC0_PAYER_ACC {
		t.Fatalf("document[0] payer account, expected %s, got %s", TEST_DOC0_PAYER_ACC, doc.PayerAccount)
	}
	if doc.PayerInn != TEST_DOC0_PAYER_INN {
		t.Fatalf("document[0] payer inn, expected %s, got %s", TEST_DOC0_PAYER_INN, doc.PayerInn)
	}
	if doc.PayerName != TEST_DOC0_PAYER_NAME {
		t.Fatalf("document[0] payer name, expected %s, got %s", TEST_DOC0_PAYER_NAME, doc.PayerName)
	}
	if doc.ReceiverInn != TEST_DOC0_REC_INN {
		t.Fatalf("document[0] receiver inn, expected %s, got %s", TEST_DOC0_REC_INN, doc.ReceiverInn)
	}
	if doc.ReceiverAccount != TEST_DOC0_REC_ACC {
		t.Fatalf("document[0] receiver account, expected %s, got %s", TEST_DOC0_REC_ACC, doc.ReceiverAccount)
	}
	if doc.Sum != TEST_DOC0_SUM {
		t.Fatalf("document[0] sum, expected %f, got %f", TEST_DOC0_SUM, doc.Sum)
	}

	doc1, ok := imp.Documents[1].(*PPDocument)
	doc2, ok := imp.Documents[2].(*PPDocument)
	// fmt.Printf("doc0:%v\n", doc)
	// fmt.Printf("doc1:%v\n", doc1)
	// fmt.Printf("doc2:%v\n", doc2)
	if !ok {
		t.Fatal("document[1] must be of type PPDocument")
	}
	if doc1.PayerAccount != TEST_DOC1_PAYER_ACC {
		t.Fatalf("document[1] payer account, expected %s, got %s", TEST_DOC1_PAYER_ACC, doc1.PayerAccount)
	}
	if doc1.PayerInn != TEST_DOC1_PAYER_INN {
		t.Fatalf("document[1] payer inn, expected %s, got %s", TEST_DOC1_PAYER_INN, doc1.PayerInn)
	}
	if doc1.PayerName != TEST_DOC1_PAYER_NAME {
		t.Fatalf("document[1] payer name, expected %s, got %s", TEST_DOC1_PAYER_NAME, doc1.PayerName)
	}
	if doc1.ReceiverInn != TEST_DOC1_REC_INN {
		t.Fatalf("document[1] receiver inn, expected %s, got %s", TEST_DOC1_REC_INN, doc1.ReceiverInn)
	}
	if doc1.ReceiverAccount != TEST_DOC1_REC_ACC {
		t.Fatalf("document[1] receiver account, expected %s, got %s", TEST_DOC1_REC_ACC, doc1.ReceiverAccount)
	}
	if doc1.Sum != TEST_DOC1_SUM {
		t.Fatalf("document[1] sum, expected %f, got %f", TEST_DOC1_SUM, doc1.Sum)
	}

	if !ok {
		t.Fatal("document[2] must be of type PPDocument")
	}
	if doc2.PayerAccount != TEST_DOC2_PAYER_ACC {
		t.Fatalf("document[2] payer account, expected %s, got %s", TEST_DOC2_PAYER_ACC, doc2.PayerAccount)
	}
	if doc2.PayerInn != TEST_DOC2_PAYER_INN {
		t.Fatalf("document[2] payer inn, expected %s, got %s", TEST_DOC2_PAYER_INN, doc2.PayerInn)
	}
	if doc2.PayerName != TEST_DOC2_PAYER_NAME {
		t.Fatalf("document[2] payer name, expected %s, got %s", TEST_DOC2_PAYER_NAME, doc2.PayerName)
	}
	if doc2.ReceiverInn != TEST_DOC2_REC_INN {
		t.Fatalf("document[2] receiver inn, expected %s, got %s", TEST_DOC2_REC_INN, doc2.ReceiverInn)
	}
	if doc2.ReceiverAccount != TEST_DOC2_REC_ACC {
		t.Fatalf("document[2] receiver account, expected %s, got %s", TEST_DOC2_REC_ACC, doc2.ReceiverAccount)
	}
	if doc2.Sum != TEST_DOC2_SUM {
		t.Fatalf("document[2] sum, expected %f, got %f", TEST_DOC2_SUM, doc2.Sum)
	}
}
