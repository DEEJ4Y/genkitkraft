package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
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

func toProviderResponse(p *provider.Provider) gen.ModelsProviderResponse {
	resp := gen.ModelsProviderResponse{
		Id:           p.ID,
		Name:         p.Name,
		ProviderType: gen.ModelsProviderType(p.ProviderType),
		ApiKey:       p.MaskedAPIKey(),
		Enabled:      p.Enabled,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.BaseURL != "" {
		resp.BaseUrl = &p.BaseURL
	}
	// Include config if non-empty
	if len(p.RawConfig) > 0 && string(p.RawConfig) != "{}" {
		var configMap map[string]string
		if err := json.Unmarshal(p.RawConfig, &configMap); err == nil && len(configMap) > 0 {
			resp.Config = &configMap
		}
	}
	return resp
}

func toProviderListResponse(result queries.ListProvidersResult) gen.ModelsProviderListResponse {
	providers := make([]gen.ModelsProviderResponse, len(result.Providers))
	for i, p := range result.Providers {
		providers[i] = toProviderResponse(p)
	}
	return gen.ModelsProviderListResponse{Providers: providers}
}

func toCreateProviderParams(req gen.ModelsCreateProviderRequest) commands.CreateProviderParams {
	params := commands.CreateProviderParams{
		Name:         req.Name,
		ProviderType: provider.ProviderType(req.ProviderType),
		APIKey:       req.ApiKey,
	}
	if req.BaseUrl != nil {
		params.BaseURL = *req.BaseUrl
	}
	if req.Config != nil {
		params.Config = *req.Config
	}
	return params
}

func toUpdateProviderParams(id string, req gen.ModelsUpdateProviderRequest) commands.UpdateProviderParams {
	return commands.UpdateProviderParams{
		ID:      id,
		Name:    req.Name,
		APIKey:  req.ApiKey,
		BaseURL: req.BaseUrl,
		Config:  req.Config,
		Enabled: req.Enabled,
	}
}

func toProviderTypeListResponse(result queries.ListProviderTypesResult) gen.ModelsProviderTypeListResponse {
	types := make([]gen.ModelsProviderTypeInfo, len(result.ProviderTypes))
	for i, pt := range result.ProviderTypes {
		info := gen.ModelsProviderTypeInfo{
			Type:            gen.ModelsProviderType(pt.Type),
			DisplayName:     pt.DisplayName,
			RequiresApiKey:  pt.RequiresAPIKey,
			RequiresBaseUrl: pt.RequiresBaseURL,
			EnvVarHint:      pt.EnvVarHint,
			ModelPrefix:     pt.ModelPrefix,
		}
		if pt.BaseURLDefault != "" {
			info.BaseUrlDefault = &pt.BaseURLDefault
		}
		if pt.ComingSoon {
			comingSoon := true
			info.ComingSoon = &comingSoon
		}
		fields := make([]gen.ModelsConfigFieldInfo, len(pt.ConfigFields))
		for j, f := range pt.ConfigFields {
			field := gen.ModelsConfigFieldInfo{
				Name:     f.Name,
				Label:    f.Label,
				Required: f.Required,
			}
			if f.Placeholder != "" {
				field.Placeholder = &f.Placeholder
			}
			if f.Sensitive {
				sensitive := true
				field.Sensitive = &sensitive
			}
			fields[j] = field
		}
		info.ConfigFields = fields
		types[i] = info
	}
	return gen.ModelsProviderTypeListResponse{ProviderTypes: types}
}
