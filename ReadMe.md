# Обмен данными с клиент-банком через формат обмена [1CClientBankExchange](https://v8.1c.ru/tekhnologii/obmen-dannymi-i-integratsiya/standarty-i-formaty/standart-obmena-s-sistemami-klient-banka/formaty-obmena/).
Реализует экспорт платежных поручений, импорт выписок по расчетным счетам. 

### Как использовать

#### Для экспорта документов в банк: 
```go
	import (
		"os"
		"time"
		
		"github.com/dronm/clbnk"
	)
	//список документов
	documents := []clbnk.BankExportDocument{&clbnk.PPDocument{Num: 1,
		Date:                time.Now(),
		Sum:                 175000,
		PayerName:           `ООО "Рога и Копыта"`,
		PayerInn:            "1234567891",
		PayerAccount:        "12345678901234567890",
		PayerBankName:       "КакойТоБанк ОАО",
		PayerBankPlace:      "г. Москва",
		PayerBankBik:        "123456789",
		PayerBankAccount:    "12345678901234567890",
		ReceiverName:        `ИП Иванов А.А.`,
		ReceiverInn:         "111122223344",
		ReceiverAccount:     "12345678901234567890",
		ReceiverBankName:    "КакойтоБанк ОАО",
		ReceiverBankPlace:   "г. Москва",
		ReceiverBankBik:     "123456789",
		ReceiverBankAccount: "12345678901234567890",
		PayType:             clbnk.PAY_TYPE_DIG,
		OplType:             "01",
		Order:               5,
		PayComment:          "За товары, по счету №125 на сумму 175000-00",
	},
		&clbnk.PPDocument{Num: 2,
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
			PayType:          clbnk.PAY_TYPE_DIG,
			OplType:          "01",
			Order:            5,
			PayComment:       "За товары, по счету №777 на сумму 375-25\nВ том числе НДС (20%) 62-54",
		},
	}
	//объект выгрузки
	exp := clbnkh.NewBankExport(documents)
	exp.EncodingType = ENCODING_TYPE_WIN
	
	//export to byte slice
	bData, err := exp.Marshal()
	if err != nil {
		panic("clbnk.Marshal() failed: %v", err)
	}	

	//сохраним в файл
	f, err := os.Create("cl_to_bank.txt")
	if err != nil {
		panic("os.Create() failed: %v", err)
	}
	defer f.Close()
	f.Write(bData)	
```
	
#### Для импорта выписок из файла банка: 
	fileCont, err := os.ReadFile("kl_to_1c.txt")
	if err != nil {
		panic(err)
	}
	imp := clbnk.NewBankImport()
	if err := imp.Unmarshal(fileCont); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
```

