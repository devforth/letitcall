package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/letitcall/letitcall/api/internal/content"
)

func (s *Server) getBranding(w http.ResponseWriter, _ *http.Request) {
	branding, err := s.store.GetBranding()
	if err != nil {
		internalError(w, err, "load branding")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"branding": branding})
}

type updateBrandingRequest struct {
	Name string  `json:"name"`
	Logo *string `json:"logo"`
}

func (s *Server) updateBranding(w http.ResponseWriter, r *http.Request) {
	var request updateBrandingRequest
	if decodeJSON(w, r, &request) != nil {
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	branding, err := s.store.GetBranding()
	if err != nil {
		internalError(w, err, "load branding")
		return
	}
	previousBranding := branding
	previousLogoFilename := branding.LogoPath
	branding.Name = request.Name
	var logo content.Logo
	if request.Logo != nil {
		logo, err = s.logos.Prepare(*request.Logo)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		branding.LogoPath = logo.Filename
		if err := s.logos.Write(logo); err != nil {
			internalError(w, err, "store brand logo")
			return
		}
	}
	if err := s.store.PutBranding(branding); err != nil {
		if logo.Filename != "" {
			_ = s.logos.Remove(logo.Filename)
		}
		internalError(w, err, "store branding")
		return
	}
	changes, err := auditDiff(previousBranding, branding)
	if err != nil {
		internalError(w, err, "build branding audit payload")
		return
	}
	if err := s.recordAuditLog(r, "edited", "branding", "current", changes); err != nil {
		internalError(w, err, "record branding audit log")
		return
	}
	if logo.Filename != "" && previousLogoFilename != "" {
		if err := s.logos.Remove(previousLogoFilename); err != nil {
			slog.Error("remove previous brand logo", "error", err, "filename", previousLogoFilename)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"branding": branding})
}
