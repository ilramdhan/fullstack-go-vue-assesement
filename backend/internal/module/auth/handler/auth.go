package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	authUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

type AuthHandler struct {
	authUC authUsecase.AuthUsecase
}

func NewAuthHandler(authUC authUsecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUC: authUC,
	}
}

func (a *AuthHandler) PostDashboardV1AuthLogin(w http.ResponseWriter, r *http.Request) {
	var req openapigen.PostDashboardV1AuthLoginJSONBody
	if !decodeJSONBody(w, r, &req) {
		return
	}
	token, user, err := a.authUC.Login(req.Email, req.Password)
	if err != nil {
		transport.WriteError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(openapigen.LoginResponse{Email: &user.Email, Role: &user.Role, Token: &token})
	if err != nil {
		transport.WriteAppError(w, entity.ErrorInternal("internal server error"))
		return
	}
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
