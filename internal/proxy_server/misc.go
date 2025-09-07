package proxy_server

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/terricain/aws_ecr_proxy/internal/version"
	"net/http"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (d *WebData) Readyz(w http.ResponseWriter, r *http.Request) {
	if len(d.fetcher.Token) == 0 || len(d.fetcher.Endpoint) == 0 {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}
}

func Version(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"version":    version.VERSION,
		"build_date": version.BUILDDATE,
		"sha":        version.SHA,
	}

	if jsonBytes, err := json.Marshal(data); err != nil {
		log.Error().Err(err).Msg("Failed to convert version info to JSON")
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
		_, _ = w.Write(jsonBytes)
	}
}
