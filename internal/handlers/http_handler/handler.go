package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

const sessionCookieName = "session_token"
const sessionMaxAge = 86400 // 24 hours

// Compile-time check that Handler implements the generated ServerInterface.
var _ gen.ServerInterface = (*Handler)(nil)

// Handler implements gen.ServerInterface, delegating to the application layer.
type Handler struct {
	app *app.AuthApp
}

func NewHandler(authApp *app.AuthApp) *Handler {
	return &Handler{app: authApp}
}

func (h *Handler) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.Queries.GetAuthStatus.Execute(r.Context(), queries.GetAuthStatusParams{})
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAuthStatusResponse(result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req gen.ModelsLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		writeAppError(w, errors.NewAppError(errors.InvalidInput, "username and password are required"))
		return
	}

	params := toLoginParams(req, extractIP(r))
	result, err := h.app.Commands.Login.Execute(r.Context(), params)
	if err != nil {
		writeAppError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    result.Token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   isSecure(r),
	})

	writeJSON(w, http.StatusOK, toLoginResponse(result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		_ = h.app.Commands.Logout.Execute(r.Context(), commands.LogoutParams{Token: cookie.Value})
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   isSecure(r),
	})

	writeJSON(w, http.StatusOK, gen.ModelsLogoutResponse{Ok: true})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeAppError(w, errors.NewAppError(errors.Unauthorized, "unauthorized"))
		return
	}

	result, err := h.app.Queries.GetMe.Execute(r.Context(), queries.GetMeParams{Token: cookie.Value})
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toMeResponse(result))
}

func (h *Handler) HealthLiveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gen.ModelsHealthCheckResponse{
		Status:  gen.Up,
		Message: "Server is running",
	})
}

func (h *Handler) HealthReadiness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, gen.ModelsHealthCheckResponse{
		Status:  gen.Up,
		Message: "Server is ready",
	})
}
