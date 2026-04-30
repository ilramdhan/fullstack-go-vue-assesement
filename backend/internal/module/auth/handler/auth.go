package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	authUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type AuthHandler struct {
	authUC authUsecase.AuthUsecase
}

func NewAuthHandler(authUC authUsecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req openapigen.LoginRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	token, user, err := h.authUC.Login(string(req.Email), req.Password)
	if err != nil {
		// Login should not leak whether the email exists or the password was wrong.
		var appErr *entity.AppError
		if errors.As(err, &appErr) && appErr.Code == entity.ErrorCodeNotFound {
			err = entity.ErrorUnauthorized("invalid credentials")
		}
		transport.WriteError(w, err)
		return
	}

	transport.WriteJSON(w, http.StatusOK, openapigen.LoginResponse{
		Token: token,
		User: openapigen.User{
			Email: openapi_types.Email(user.Email),
			Role:  openapigen.UserRole(user.Role),
		},
	})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil {
		transport.WriteAppError(w, entity.ErrorBadRequest("empty body"))
		return false
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		transport.WriteAppError(w, entity.ErrorBadRequest("failed to read body"))
		return false
	}
	if err := json.Unmarshal(body, dst); err != nil {
		transport.WriteAppError(w, entity.ErrorBadRequest("invalid json: "+err.Error()))
		return false
	}
	return true
}
