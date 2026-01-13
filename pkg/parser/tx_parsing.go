package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/user/go-frontol-loader/pkg/models"
)

type ParsedTransaction struct {
	Table string
	Value interface{}
}

func wrapParsed[T any](fields []string, sourceFolder string, tableName string) (ParsedTransaction, error) {
	val, err := parseTxModel[T](fields, sourceFolder, tableName)
	if err != nil {
		return ParsedTransaction{}, err
	}
	return ParsedTransaction{Table: tableName, Value: val}, nil
}

func appendTx[T any](table string, value interface{}, dst *[]T) error {
	typed, ok := value.(T)
	if !ok {
		return fmt.Errorf("invalid value for %s: %T", table, value)
	}
	*dst = append(*dst, typed)
	return nil
}

func parseTxModel[T any](fields []string, sourceFolder string, tableName string) (T, error) {
	var dest T
	schema, ok := models.TxSchemas[tableName]
	if !ok {
		return dest, fmt.Errorf("unknown tx schema: %s", tableName)
	}
	if err := fillTxStruct(&dest, fields, sourceFolder, schema); err != nil {
		return dest, err
	}
	return dest, nil
}

func fillTxStruct(dst interface{}, fields []string, sourceFolder string, schema []models.TxColumnSpec) error {
	val := reflect.ValueOf(dst)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be pointer to struct")
	}
	if val.IsNil() {
		panic("destination is nil")
	}
	if val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination must be pointer to struct")
	}
	structVal := val.Elem()
	fieldIndex := 0

	for _, spec := range schema {
		fieldName := models.ColumnToFieldName(spec.Name)
		field := structVal.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("missing field %s on %s", fieldName, structVal.Type().Name())
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field %s on %s", fieldName, structVal.Type().Name())
		}

		if spec.Kind == models.TxColumnSource {
			if field.Kind() != reflect.String {
				return fmt.Errorf("source_folder field %s is not string", fieldName)
			}
			field.SetString(sourceFolder)
			continue
		}

		var raw string
		if fieldIndex < len(fields) {
			raw = fields[fieldIndex]
		}
		fieldIndex++

		switch spec.Kind {
		case models.TxColumnString:
			if field.Kind() != reflect.String {
				return fmt.Errorf("field %s is not string", fieldName)
			}
			field.SetString(raw)
		case models.TxColumnInt64:
			val := int64(0)
			if raw != "" {
				parsed, err := strconv.ParseInt(raw, 10, 64)
				if err == nil {
					val = parsed
				}
			}
			if field.Kind() != reflect.Int64 {
				return fmt.Errorf("field %s is not int64", fieldName)
			}
			field.SetInt(val)
		case models.TxColumnFloat64:
			val := float64(0)
			if raw != "" {
				parsed, err := parseFloatWithComma(raw)
				if err == nil {
					val = parsed
				}
			}
			if field.Kind() != reflect.Float64 {
				return fmt.Errorf("field %s is not float64", fieldName)
			}
			field.SetFloat(val)
		case models.TxColumnDate:
			parsed := time.Time{}
			if raw != "" {
				if t, err := time.Parse("02.01.2006", raw); err == nil {
					parsed = t
				}
			}
			if field.Type() != reflect.TypeOf(time.Time{}) {
				return fmt.Errorf("field %s is not time.Time", fieldName)
			}
			field.Set(reflect.ValueOf(parsed))
		case models.TxColumnTime:
			parsed := time.Time{}
			if raw != "" {
				if t, err := time.Parse("15:04:05", raw); err == nil {
					parsed = t
				}
			}
			if field.Type() != reflect.TypeOf(time.Time{}) {
				return fmt.Errorf("field %s is not time.Time", fieldName)
			}
			field.Set(reflect.ValueOf(parsed))
		default:
			return fmt.Errorf("unsupported column kind for %s", spec.Name)
		}
	}

	return nil
}
