package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseError struct {
	Status  int
	Message string `json:"message"`
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("status code %d: %s", e.Status, e.Message)
}

func ParseResponseError(status int, body []byte) error {
	if status < http.StatusBadRequest {
		return nil
	}

	responseError := ResponseError{
		Status: status,
	}

	if json.Valid(body) {
		if err := json.Unmarshal(body, &responseError); err == nil {
			return responseError
		}
	}

	responseError.Message = string(body)

	return responseError
}
