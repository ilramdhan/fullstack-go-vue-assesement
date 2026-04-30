package transport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
)

type errorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func codeToStatus(code entity.Code) int {
	switch code {
	case entity.ErrorCodeBadRequest:
		return http.StatusBadRequest
	case entity.ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case entity.ErrorCodeForbidden:
		return http.StatusForbidden
	case entity.ErrorCodeNotFound:
		return http.StatusNotFound
	case entity.ErrorCodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("transport: encode response: %v", err)
	}
}

func WriteAppError(w http.ResponseWriter, appErr *entity.AppError) {
	status := codeToStatus(appErr.Code)
	WriteJSON(w, status, errorBody{
		Code:    status,
		Message: appErr.Message,
		Details: appErr.Details,
	})
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var aErr *entity.AppError
	if errors.As(err, &aErr) {
		WriteAppError(w, aErr)
		return
	}
	log.Printf("transport: unexpected error: %v", err)
	WriteJSON(w, http.StatusInternalServerError, errorBody{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}
