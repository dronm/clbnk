package clbnk

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// importDocumentTypes contains all documents for import.
var importDocumentMaps = map[DocumentType]reflect.Type{DOCUMENT_TYPE_BANK_ORDER: reflect.TypeOf(BankOrderDocument{}),
	DOCUMENT_TYPE_PP: reflect.TypeOf(PPDocument{}),
}

// Some error texts.
const (
	ER_NO_ENC       = "encoding not defined"
	ER_INVALID_FILE = "invalid file format"
)

// Unmarshaler is for structures with custom unmarshal functions.
type Unmarshaler interface {
	Unmarshal(string) error
}

// Import field types: ordinary field, start of a slice elentnt,
// end of a slice element.
const (
	FIELD_TYPE_FIELD ImportFieldType = iota
	FIELD_TYPE_ELEM_START
	FIELD_TYPE_ELEM_END
)

type ImportFieldType int

// Unmarshal is the main entry point for importing data.
// It starts with identifying file encoding type and checking file header.
func (e *BankImport) Unmarshal(data []byte) error {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return nil
	}
	if len(lines) < 3 {
		return fmt.Errorf(ER_INVALID_FILE)
	}
	if lines[0] != HEADER {
		return fmt.Errorf("file header not found %v!=%v", []byte(lines[0]), []byte(HEADER))
	}
	if e.EncodingType == ENCODING_TYPE_NOT_DEFINED {
		enc := strings.Split(lines[2], "=")
		if len(enc) < 2 {
			return fmt.Errorf(ER_NO_ENC)
		}
		if enc[1] == ENCODING_WIN {
			e.EncodingType = ENCODING_TYPE_WIN

		} else if enc[1] == ENCODING_DOS {
			e.EncodingType = ENCODING_TYPE_DOS

		} else {
			return fmt.Errorf(ER_NO_ENC)
		}
	}

	//decode data from given encoding type
	data_dec, err := e.EncodingType.decode(data)
	if err != nil {
		return err
	}

	//farther we will deal with strings as struct fields are strings
	data_str := strings.ReplaceAll(string(data_dec), "\r\n", "\n")
	lines = strings.Split(data_str, "\n")
	n := 0
	res := unmarshal(lines, &n, reflect.ValueOf(e), FOOTER)
	return res
}

func unmarshal(lines []string, lineNum *int, v reflect.Value, endSection string) error { // Ensure dataPtr is a pointer to a struct
	// Dereference the pointer to get the struct value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for {
		if *lineNum >= len(lines) {
			break
		}
		line := lines[*lineNum]
		if line == "" {
			*lineNum++
			continue
		}

		var field_id string
		var field_val string

		ind := strings.Index(line, "=")
		if ind == -1 {
			field_id = line
		} else {
			field_id = line[:ind]
			field_val = line[ind+1:]
		}
		*lineNum++

		// fmt.Printf("field_id:%s, field_val=%s, endSection:%s\n", field_id, field_val, endSection)
		if field_id == endSection {
			return nil
		}
		struct_field, found, field_type, sec_end := findFieldByName(v, field_id)
		if !found {
			// fmt.Println("not found field", field_id)
			continue
		}

		fmt.Println("ID:", field_id, "VAL:", field_val)
		if err := setFieldValue(struct_field, field_val, field_type == FIELD_TYPE_ELEM_START, sec_end, lines, lineNum); err != nil {
			return err
		}
	}
	return nil
}

// findFieldByName finds the field in the struct with the specified custom tag name.
// The function returns field value if it is found, bool indicationg if field is found,
// the found field type and section end tag.
func findFieldByName(v reflect.Value, tagName string) (reflect.Value, bool, ImportFieldType, string) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)

		tag := field.Tag.Get("bank")
		elem_start := field.Tag.Get("bankElemStart")
		elem_end := field.Tag.Get("bankElemEnd")

		switch tagName {
		case tag:
			return v.Field(i), true, FIELD_TYPE_FIELD, ""
		case elem_start:
			return v.Field(i), true, FIELD_TYPE_ELEM_START, elem_end
		case elem_end:
			return v.Field(i), true, FIELD_TYPE_ELEM_END, ""
		}
	}
	return reflect.Value{}, false, FIELD_TYPE_FIELD, ""
}

// setFieldValue sets the value of the field according to its type.
func setFieldValue(field reflect.Value, value string, isElemStart bool, endSection string, lines []string, lineNum *int) error {
	// fmt.Println("fieldKind=", field.Kind(), "value=", value, "isElemStart=", isElemStart)
	if isElemStart {
		//slice element or structure elemen
		if field.Kind() == reflect.Struct {
			//structure field
			return unmarshal(lines, lineNum, field.Addr(), endSection)

		} else if field.Kind() == reflect.Slice {
			slice_elem := reflect.New(field.Type().Elem()).Elem()

			//documents
			if ok := slice_elem.Type().Implements(reflect.TypeOf((*BankImportDocument)(nil)).Elem()); ok {
				//determine document type by value
				var doc_type reflect.Type
				for i, d_tp := range DocumentTypeValues() {
					if d_tp == value {
						doc_type = importDocumentMaps[DocumentType(i)]
						break
					}
				}
				// if reflect.Zero(reflect.TypeOf(doc_type)) == doc_type {
				if doc_type == nil {
					return fmt.Errorf("document type not found by ID %s", value)
				}
				// slice_elem = reflect.ValueOf(&BankOrderDocument{})
				//create a new instance of doc_type
				slice_elem = reflect.New(doc_type)
				if err := unmarshal(lines, lineNum, slice_elem.Elem().Addr(), endSection); err != nil {
					return err
				}

			} else {
				if err := unmarshal(lines, lineNum, slice_elem, endSection); err != nil {
					return err
				}
			}
			//does not work with field.Append
			slice_len := field.Len()
			new_slice := reflect.MakeSlice(field.Type(), slice_len+1, slice_len+1)
			reflect.Copy(new_slice, field)
			new_slice.Index(slice_len).Set(slice_elem)
			field.Set(new_slice)

			return nil
		}
		return fmt.Errorf("tag 'bankElemStart' must belong to a struct or a slice")
	}

	if value == "" {
		return nil
	}
	if field.Type() == reflect.TypeOf(time.Time{}) {
		t, err := time.Parse("02.01.2006", value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(t))
		return nil
	}
	if inf, ok := field.Addr().Interface().(Unmarshaler); ok {
		return inf.Unmarshal(value)
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("Unmarshal: failed to parse int value: %v", err)
		}
		field.SetInt(int64(intValue))

	case reflect.Float32:
		fValue, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return fmt.Errorf("Unmarshal: failed to parse float32 value: %v", err)
		}
		field.SetFloat(fValue)

	case reflect.Float64:
		fValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("Unmarshal: failed to parse float64 value: %v", err)
		}
		field.SetFloat(fValue)

	default:
		return fmt.Errorf("Unmarshal: unsupported field type: %s", field.Kind())
	}
	return nil
}
