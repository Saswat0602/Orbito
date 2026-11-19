package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/orbito/api-go/internal/shared/apperror"
)

var validate = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}()

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

		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return apperror.New(http.StatusRequestEntityTooLarge, "PayloadTooLargeError", "request entity too large")
		}

		// Mirror NestJS forbidNonWhitelisted: ["property <name> should not exist"]
		errStr := err.Error()
		if strings.Contains(errStr, "unknown field") {
			parts := strings.Split(errStr, `"`)
			if len(parts) >= 2 {
				fieldName := parts[1]
				return apperror.Validation([]string{fmt.Sprintf("property %s should not exist", fieldName)})
			}
		}

		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return apperror.BadRequest("Unexpected end of JSON input")
		}

		return apperror.BadRequest(fmt.Sprintf("Invalid JSON payload: %v", err))
	}

	if err := validate.Struct(target); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			messages := make([]string, 0, len(valErrs))
			for _, fe := range valErrs {
				field := fe.Field()
				if field == "" {
					field = strings.ToLower(fe.StructField())
				}
				switch fe.Tag() {
				case "required":
					messages = append(messages, fmt.Sprintf("%s should not be empty", field))
				case "email":
					messages = append(messages, fmt.Sprintf("%s must be an email", field))
				case "min":
					messages = append(messages, fmt.Sprintf("%s must be longer than or equal to %s characters", field, fe.Param()))
				case "max":
					messages = append(messages, fmt.Sprintf("%s must be shorter than or equal to %s characters", field, fe.Param()))
				case "uuid":
					messages = append(messages, fmt.Sprintf("%s must be a UUID", field))
				default:
					messages = append(messages, fmt.Sprintf("%s failed validation on '%s'", field, fe.Tag()))
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
