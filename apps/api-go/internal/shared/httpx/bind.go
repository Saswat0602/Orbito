package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/orbito/api-go/internal/shared/apperror"
)

var validate = validator.New()

// BindJSON decodes the JSON request body into target and validates struct tags.
func BindJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return apperror.BadRequest("Missing request body")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Mirror NestJS forbidNonWhitelisted: true

	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return apperror.BadRequest("Request body must not be empty")
		}
		return apperror.BadRequest(fmt.Sprintf("Invalid JSON payload: %v", err))
	}

	if err := validate.Struct(target); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			messages := make([]string, 0, len(valErrs))
			for _, fe := range valErrs {
				field := strings.ToLower(fe.Field())
				switch fe.Tag() {
				case "required":
					messages = append(messages, fmt.Sprintf("%s must not be empty", field))
				case "email":
					messages = append(messages, fmt.Sprintf("%s must be an email", field))
				case "min":
					messages = append(messages, fmt.Sprintf("%s must be at least %s characters long", field, fe.Param()))
				case "max":
					messages = append(messages, fmt.Sprintf("%s must not exceed %s characters", field, fe.Param()))
				case "uuid":
					messages = append(messages, fmt.Sprintf("%s must be a valid UUID", field))
				default:
					messages = append(messages, fmt.Sprintf("%s failed on '%s' validation", field, fe.Tag()))
				}
			}
			return apperror.Validation(messages)
		}
		return apperror.BadRequest(err.Error())
	}

	return nil
}

// Validator returns the shared validator instance.
func Validator() *validator.Validate {
	return validate
}
