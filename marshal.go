package clbnk

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type Marshaler interface {
	Marshal() ([]byte, error)
}

type FirmMarshaler interface {
	MarshalFirmName() ([]byte, error)
}

type MarshalTime interface {
	Format(string) string
}

func marshal(data interface{}, elemStart, elemEnd string) ([]byte, error) {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Dereference the pointer
	}

	if m, ok := v.Interface().(Marshaler); ok {
		return m.Marshal()
	}

	if m, ok := v.Interface().(MarshalTime); ok {
		return []byte(m.Format("02.01.2006")), nil
	}

	// fmt.Printf("Kind=%s val: %+v\n", v.Kind(), data)
	switch v.Kind() {
	case reflect.Struct:
		return marshalStruct(v)

	case reflect.Slice:
		return marshalSlice(v, elemStart, elemEnd)

	default:
		//primitive types
		switch dt := data.(type) {
		case int:
			return []byte(fmt.Sprintf("%d", dt)), nil
		case float64:
			return []byte(fmt.Sprintf("%.2f", dt)), nil
		case string:
			return []byte(dt), nil
		}
	}

	return []byte{}, fmt.Errorf("unsupported type: %s", v.Kind())
}

// If elemStart/elemEnd defined then every slice element is
// prefixed/postfixed with elemStart/elemEnd values.
func marshalSlice(v reflect.Value, elemStart, elemEnd string) ([]byte, error) {
	var buf bytes.Buffer

	for j := 0; j < v.Len(); j++ {
		slice_elem := v.Index(j)
		cont, err := marshal(slice_elem.Interface(), "", "")
		if err != nil {
			return []byte{}, err
		}
		if elemStart != "" {
			if _, err := buf.WriteString(elemStart); err != nil {
				return []byte{}, err
			}
		}
		if _, err := buf.Write(cont); err != nil {
			return []byte{}, err
		}
		if elemEnd != "" {
			if _, err := buf.WriteString(elemEnd); err != nil {
				return []byte{}, err
			}
		}
	}
	return buf.Bytes(), nil
}

func marshalStruct(v reflect.Value) ([]byte, error) {
	var buf bytes.Buffer
	var firm_name []byte
	if firm_m, ok := v.Interface().(FirmMarshaler); ok {
		var err error
		firm_name, err = firm_m.MarshalFirmName()
		if err != nil {
			return []byte{}, err
		}
	}
	// Iterate over struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		field_name := field.Tag.Get("bank")
		field_is_firm := field.Tag.Get("bankFirmName")

		field_elem_start := field.Tag.Get("bankElemStart")
		field_elem_end := field.Tag.Get("bankElemEnd")

		var field_val []byte
		if field_is_firm == "1" {
			field_val = firm_name
		} else {
			var err error
			field_val, err = marshal(v.Field(i).Interface(), field_elem_start, field_elem_end)
			if err != nil {
				return []byte{}, err
			}
		}

		comment_lines := field.Tag.Get("lines")
		if comment_lines == "" || len(field_val) == 0 { // one line value
			b := marshalField(field_name, field_val)
			if _, err := buf.Write(b); err != nil {
				return []byte{}, err
			}

		} else { // multyline value
			n, err := strconv.Atoi(comment_lines)
			if err != nil {
				return []byte{}, err
			}
			last_line := make([]byte, 0)
			lines := bytes.Split(field_val, []byte("\n"))
			b := marshalField(field_name, bytes.Join(lines, []byte{' '})) //all lines with a space
			if _, err := buf.Write(b); err != nil {
				return []byte{}, err
			}
			for j, l := range lines {
				if j+1 >= n {
					last_line = append(last_line, ' ')
					last_line = append(last_line, l...)
					continue
				}
				b := marshalField(fmt.Sprintf("%s%d", field_name, j+1), l)
				if _, err := buf.Write(b); err != nil {
					return []byte{}, err
				}
			}
			if len(last_line) > 0 {
				b := marshalField(fmt.Sprintf("%s%d", field_name, n), last_line)
				if _, err := buf.Write(b); err != nil {
					return []byte{}, err
				}

			}
		}
	}
	return buf.Bytes(), nil
}

// marshals field name=value\n
func marshalField(fieldName string, fieldVal []byte) []byte {
	if fieldName == "" && len(fieldVal) == 0 {
		return []byte{}

	} else if fieldName != "" {
		b := make([]byte, 0)
		b = append(b, []byte(fieldName+"=")...)
		b = append(b, fieldVal...)
		b = append(b, []byte("\r\n")...)
		return b
	}
	return fieldVal
}

// custom marshalling of a firm name
func marshalFirmName(inn, name string) []byte {
	//ИНН(3 chars) + space + p.Inn + space + p.Name
	b := make([]byte, 3+1+len(inn)+1+len(name))
	copy(b, []byte("ИНН "+inn+" "+name))
	return b
}
