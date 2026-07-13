package httpapi

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/swaggest/swgui"
	swagger "github.com/swaggest/swgui/v5"
)

//go:embed openapi.json
var openAPITemplate []byte

func (s *Server) openAPISpecification(w http.ResponseWriter, _ *http.Request) {
	encodedBaseURL, _ := json.Marshal(s.compatibilityBaseURL())
	baseURL := strings.Trim(string(encodedBaseURL), "\"")
	specification := strings.ReplaceAll(string(openAPITemplate), "__BASE_URL__", baseURL)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(specification))
}

func (s *Server) swaggerHandler() http.Handler {
	externalPath := s.cfg.HTTP.BasePath() + compatibilityAPIPath + "/swagger"
	return swagger.NewHandlerWithConfig(swgui.Config{
		Title:            "Let It Call Lead Generation API",
		SwaggerJSON:      s.compatibilityBaseURL() + "/openapi.json",
		BasePath:         externalPath,
		InternalBasePath: compatibilityAPIPath + "/swagger",
		ShowTopBar:       true,
	})
}
