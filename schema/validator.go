package schema

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const EMAIL_REGEX = `^.+@.+\..+$`

var validate *validator.Validate

func Init() {
	validate = validator.New()
	validate.RegisterValidation("email", validateEmail)
}

func validateEmail(field validator.FieldLevel) bool {
	return regexp.MustCompile(EMAIL_REGEX).MatchString(field.Field().String())
}
