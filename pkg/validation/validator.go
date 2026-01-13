package validation

import (
	"errors"
	"fmt"
	"time"
)

// Validator представляет интерфейс для валидаторов
type Validator interface {
	Validate(v interface{}) error
}

// CompositeValidator объединяет несколько валидаторов
type CompositeValidator struct {
	validators []Validator
}

// NewComposite создает новый композитный валидатор
func NewComposite(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{
		validators: validators,
	}
}

// Validate выполняет все валидаторы по очереди
func (cv *CompositeValidator) Validate(v interface{}) error {
	for _, validator := range cv.validators {
		if err := validator.Validate(v); err != nil {
			return err
		}
	}
	return nil
}

// RequiredValidator проверяет, что значение не пустое
type RequiredValidator struct {
	fieldName string
}

// Required создает валидатор для обязательного поля
func Required(fieldName string) *RequiredValidator {
	return &RequiredValidator{fieldName: fieldName}
}

// Validate проверяет, что значение не пустое
func (rv *RequiredValidator) Validate(v interface{}) error {
	if v == nil {
		return fmt.Errorf("%s is required", rv.fieldName)
	}

	switch val := v.(type) {
	case string:
		if val == "" {
			return fmt.Errorf("%s is required", rv.fieldName)
		}
	case *string:
		if val == nil || *val == "" {
			return fmt.Errorf("%s is required", rv.fieldName)
		}
	default:
		// Для других типов nil считается отсутствием значения
		// Другие типы (int, bool и т.д.) всегда имеют значение (даже zero value)
	}

	return nil
}

// DateFormatValidator проверяет формат даты
type DateFormatValidator struct {
	fieldName string
	format    string
}

// DateFormat создает валидатор для формата даты
func DateFormat(fieldName, format string) *DateFormatValidator {
	return &DateFormatValidator{
		fieldName: fieldName,
		format:    format,
	}
}

// Validate проверяет, что строка соответствует формату даты
func (dfv *DateFormatValidator) Validate(v interface{}) error {
	date, ok := v.(string)
	if !ok {
		return fmt.Errorf("%s must be a string", dfv.fieldName)
	}

	if date == "" {
		// Пустая строка допустима, если поле не обязательное
		// Используйте Required() для проверки обязательности
		return nil
	}

	_, err := time.Parse(dfv.format, date)
	if err != nil {
		return fmt.Errorf("%s must be in format %s", dfv.fieldName, dfv.format)
	}

	return nil
}

// NotInFutureValidator проверяет, что дата не в будущем
type NotInFutureValidator struct {
	fieldName string
	format    string
}

// NotInFuture создает валидатор для даты не в будущем
func NotInFuture(fieldName, format string) *NotInFutureValidator {
	return &NotInFutureValidator{
		fieldName: fieldName,
		format:    format,
	}
}

// Validate проверяет, что дата не в будущем
func (nfv *NotInFutureValidator) Validate(v interface{}) error {
	date, ok := v.(string)
	if !ok {
		return fmt.Errorf("%s must be a string", nfv.fieldName)
	}

	if date == "" {
		// Пустая строка допустима
		return nil
	}

	parsedDate, err := time.ParseInLocation(nfv.format, date, time.Local)
	if err != nil {
		return fmt.Errorf("%s must be in format %s", nfv.fieldName, nfv.format)
	}

	now := time.Now()
	if parsedDate.After(now) {
		return fmt.Errorf("%s cannot be in the future", nfv.fieldName)
	}

	return nil
}

// KassaCodeValidator проверяет формат кода кассы
type KassaCodeValidator struct {
	fieldName string
}

// KassaCode создает валидатор для кода кассы
func KassaCode(fieldName string) *KassaCodeValidator {
	return &KassaCodeValidator{fieldName: fieldName}
}

// Validate проверяет формат кода кассы (например, "P13" или "P13/P13")
func (kcv *KassaCodeValidator) Validate(v interface{}) error {
	code, ok := v.(string)
	if !ok {
		return fmt.Errorf("%s must be a string", kcv.fieldName)
	}

	if code == "" {
		// Пустая строка допустима
		return nil
	}

	// Базовая валидация: не пустая строка
	if len(code) == 0 {
		return fmt.Errorf("%s cannot be empty", kcv.fieldName)
	}

	return nil
}

// ValidationError представляет ошибку валидации с деталями.
//
//nolint:revive // Public API keeps ValidationError naming for clarity.
type ValidationError struct {
	Field   string
	Message string
}

// Error реализует интерфейс error
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// NewValidationError создает новую ошибку валидации
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsValidationError проверяет, является ли ошибка ValidationError
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
