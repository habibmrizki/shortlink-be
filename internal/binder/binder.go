package binder

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) map[string]string {
	out := make(map[string]string)

	if validationError, ok := errors.AsType[validator.ValidationErrors](err); ok {
		for _, e := range validationError {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			switch tag {
			case "required":
				out[field] = fmt.Sprintf("The %s field is required.", field)
			case "email":
				out[field] = fmt.Sprintf("The %s field is not a valid email address.", field)
			case "url":
				out[field] = fmt.Sprintf("The %s field is not a valid URL.", field)
			case "gt":
				out[field] = fmt.Sprintf("The %s field must be greater than %s.", field, param)
			case "gte":
				out[field] = fmt.Sprintf("The %s field must be greater than or equal to %s.", field, param)
			case "lte":
				out[field] = fmt.Sprintf("The %s field must be less than or equal to %s.", field, param)
			case "min":
				out[field] = fmt.Sprintf("The %s field must be at least %s characters long.", field, param)
			case "max":
				out[field] = fmt.Sprintf("The %s field must be at most %s characters long.", field, param)
			case "len":
				out[field] = fmt.Sprintf("The %s field must be exactly %s characters long.", field, param)
			case "numeric":
				out[field] = fmt.Sprintf("The %s field must contain only numeric characters.", field)
			case "oneof":
				out[field] = fmt.Sprintf("The %s field must be one of the following: %s.", field, param)
			case "image_max_size":
				paramValue, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					out[field] = fmt.Sprintf("The %s field has an invalid size parameter.", field)
					continue
				}
				out[field] = fmt.Sprintf("The %s field must not exceed %d MB in size.", field, int(paramValue/(1024*1024)))
			case "image_type":
				out[field] = fmt.Sprintf("The %s field must be a valid image file (JPEG, PNG, BMP, or HEIC).", field)
			case "shortlink_slug":
				out[field] = fmt.Sprintf("The %s field must be 3-50 characters, contain only letters, numbers, and hyphens, and not be a reserved word.", field)
			default:
				out[field] = fmt.Sprintf("The %s field does not meet the validation requirements for %s %s.", field, tag, param)
			}
		}
		return out
	}

	log.Printf("Unexpected error: %v", err)
	out["error"] = "An unexpected error occurred during validation."
	return out
}

