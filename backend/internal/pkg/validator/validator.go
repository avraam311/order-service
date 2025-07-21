package validator

import "github.com/go-playground/validator/v10"

type GoValidator struct {
	validate *validator.Validate
}

func New() *GoValidator {
	return &GoValidator{
		validate: validator.New(),
	}
}

func (v *GoValidator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}
