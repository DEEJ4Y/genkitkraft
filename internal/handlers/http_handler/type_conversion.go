package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

func toLoginParams(req gen.ModelsLoginRequest, clientIP string) commands.LoginParams {
	return commands.LoginParams{
		Username: req.Username,
		Password: req.Password,
		ClientIP: clientIP,
	}
}

func toLoginResponse(result commands.LoginResult) gen.ModelsLoginResponse {
	return gen.ModelsLoginResponse{Username: result.Username}
}

func toMeResponse(result queries.GetMeResult) gen.ModelsMeResponse {
	return gen.ModelsMeResponse{Username: result.Username}
}

func toAuthStatusResponse(result queries.GetAuthStatusResult) gen.ModelsAuthStatusResponse {
	return gen.ModelsAuthStatusResponse{Required: result.Required}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeAppError(w http.ResponseWriter, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		writeJSON(w, errors.HTTPStatusCode(appErr.Code()), gen.ModelsErrorResponse{Error: appErr.Error()})
		return
	}
	writeJSON(w, http.StatusInternalServerError, gen.ModelsErrorResponse{Error: "internal server error"})
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.SplitN(fwd, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func isSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}
