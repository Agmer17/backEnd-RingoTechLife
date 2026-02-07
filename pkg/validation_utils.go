package pkg

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return "terlalu pendek"
	case "max":
		return "terlalu panjang"
	case "numeric":
		return "harus berupa angka"
	default:
		return "tidak valid"
	}
}

func PhoneID(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return strings.HasPrefix(phone, "08") && len(phone) >= 10
}

func ValidationErrorsToMap(err error) map[string]string {
	errors := make(map[string]string)

	if err == nil {
		return errors
	}

	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field() // nama field struct
		errors[field] = validationMessage(e)
	}

	return errors
}
