package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("failed to write json response", "error", err)
		}
	}
}

func Success(w http.ResponseWriter, message string, data any) {
	JSON(w, http.StatusOK, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(w http.ResponseWriter, message string, data any) {
	JSON(w, http.StatusCreated, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, status int, message string, errDetails any) {
	JSON(w, status, Envelope{
		Success: false,
		Message: message,
		Errors:  errDetails,
	})
}

func BadRequest(w http.ResponseWriter, message string, errDetails any) {
	Error(w, http.StatusBadRequest, message, errDetails)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, nil)
}

func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message, nil)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, nil)
}

func InternalServerError(w http.ResponseWriter, err error) {
	if err != nil {
		slog.Error("internal server error", "error", err)
	}
	Error(w, http.StatusInternalServerError, "An unexpected error occurred", nil)
}
