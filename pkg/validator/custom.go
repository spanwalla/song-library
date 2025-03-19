package validator

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

type CustomValidator struct {
	v *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	cv := &CustomValidator{v: v}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := v.RegisterValidation("date", cv.validateDate)
	if err != nil {
		panic(err)
	}

	return cv
}

func (cv *CustomValidator) Validate(i any) error {
	err := cv.v.Struct(i)
	if err != nil {
		fieldErr := err.(validator.ValidationErrors)[0]

		return cv.newValidationError(fieldErr.Field(), fieldErr.Tag(), fieldErr.Param())
	}
	return nil
}

func (cv *CustomValidator) newValidationError(field string, tag string, param string) error {
	switch tag {
	case "required":
		return fmt.Errorf("field %s is required", field)
	case "min":
		return fmt.Errorf("field %s must be at least %s characters", field, param)
	case "max":
		return fmt.Errorf("field %s must be at most %s characters", field, param)
	case "date":
		return fmt.Errorf("field %s must be a valid date (format: 2006-01-17)", field)
	case "uri":
		return fmt.Errorf("field %s must be a valid URI", field)
	case "number":
		return fmt.Errorf("field %s must be a valid number", field)
	case "gt":
		return fmt.Errorf("field %s must be greater than %s", field, param)
	default:
		return fmt.Errorf("field %s is invalid", field)
	}
}

func (cv *CustomValidator) validateDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}
