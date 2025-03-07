package error

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ValidationResponse struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

var ErrValidator = map[string]string{}

func ErrValidationResponse(err error) (validationResponses []ValidationResponse) {
	var fieldErrors validator.ValidationErrors
	if errors.As(err, &fieldErrors) {
		for _, err := range fieldErrors {
			switch err.Tag() {
			case "required":
				validationResponses = append(validationResponses, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s is a required field", err.Field()),
				})
			case "email":
				validationResponses = append(validationResponses, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s must be a valid email", err.Field()),
				})
			case "oneof":
				validationResponses = append(validationResponses, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s must be an oneof [%s]", err.Field(), err.Param()),
				})
			default:
				errValidator, ok := ErrValidator[err.Tag()]
				if ok {
					count := strings.Count(errValidator, "%s")
					if count == 1 {
						validationResponses = append(validationResponses, ValidationResponse{
							Field:   err.Field(),
							Message: fmt.Sprintf(errValidator, err.Field()),
						})
					} else {
						validationResponses = append(validationResponses, ValidationResponse{
							Field:   err.Field(),
							Message: fmt.Sprintf(errValidator, err.Field(), err.Param()),
						})
					}
				} else {
					validationResponses = append(validationResponses, ValidationResponse{
						Field:   err.Field(),
						Message: fmt.Sprintf("something wrong on %s; %s", err.Field(), err.Tag()),
					})
				}
			}
		}
	}
	return validationResponses
}

func WrapError(err error) error {
	log.Errorf("error: %+v\n", err)
	return err
}
