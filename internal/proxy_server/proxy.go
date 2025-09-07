package proxy_server

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/peterhellberg/link"
	"github.com/rs/zerolog/log"
	"github.com/terricain/aws_ecr_proxy/internal/ecr_token"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type WebData struct {
	fetcher    *ecr_token.EcrFetcher
	httpClient *http.Client
}

func (d *WebData) Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if len(d.fetcher.Token) == 0 || len(d.fetcher.Endpoint) == 0 {
		log.Error().Msg("ECR token missing, this is generally not good")
		w.WriteHeader(500)
		return
	}

	// Endpoint does not end with a trailing slash and r.RequestURI starts with a slash and contains querystrings
	ecsUrl := d.fetcher.Endpoint + r.RequestURI

	// Create new request
	req, err := http.NewRequest(r.Method, ecsUrl, r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create an outbound request object")
		w.WriteHeader(500)
		return
	}
	// Copy current request headers to the outgoing request
	d.CopyHeaders(r, req)

	// Do request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to submit the outbound request object")
		w.WriteHeader(500)
		return
	}

	// Copy response headers to our http response
	// Do header dodgyness, basically the link header can contain a link to ECR with a querystring
	// so we need to replace the scheme and host of that url with ourselves
	for name := range resp.Header {
		if name == "Link" {
			headerValue := resp.Header.Get(name)
			header, err := FixLinkHeader(r.URL.Scheme, r.Host, headerValue)
			if err != nil {
				log.Error().Err(err).Str("LinkHeader", headerValue).Msg("Failed to parse and fudge a Link header")
				w.WriteHeader(500)
				return
			}
			w.Header().Set(name, header)
		} else {
			w.Header().Set(name, resp.Header.Get(name))
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy response body to http response
	size, _ := io.Copy(w, resp.Body)
	dur := time.Since(start)

	log.Info().Str("method", r.Method).Stringer("url", r.URL).Int("status", resp.StatusCode).Int64("size", size).Dur("duration_ms", dur).Msg("")
}

func (d *WebData) CopyHeaders(src *http.Request, dst *http.Request) {
	for name, values := range src.Header {
		if name == "Authorization" {
			continue
		}

		for _, value := range values {
			dst.Header.Add(name, value)
		}

	}
	dst.Header.Add("Authorization", "Basic "+d.fetcher.Token)
}

func FixLinkHeader(scheme, host, header string) (string, error) {
	if scheme == "" {
		scheme = "http"
	}

	linkHdr := ""

	index := 0
	// Parse link header into its component parts
	for _, linkDetail := range link.Parse(header) {
		if index > 0 {
			linkHdr += ", "
		}

		// Parse the link url
		urlPart, err := url.Parse(linkDetail.URI)
		if err != nil {
			return "", err
		}

		// Create a new url from various sources
		newUrlPart := url.URL{
			Scheme:   scheme,
			Host:     host,
			Path:     urlPart.Path,
			RawQuery: urlPart.RawQuery,
		}

		// Link library doesnt contain method to generate a header so we do that here
		linkHdr += "<" + newUrlPart.String() + ">"
		if len(linkDetail.Rel) > 0 {
			linkHdr += "; rel=\"" + linkDetail.Rel + "\""
		}
		for key, value := range linkDetail.Extra {
			linkHdr += "; " + key + "=\"" + value + "\""
		}

		index += 1
	}

	return linkHdr, nil
}

func Run(addr string, disableProxyHeaders bool, fetcher *ecr_token.EcrFetcher) {

	webHandlers := WebData{
		fetcher: fetcher,
		httpClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	r := mux.NewRouter()
	if !disableProxyHeaders {
		r.Use(handlers.ProxyHeaders)
	}

	r.HandleFunc("/healthz", Healthz)
	r.HandleFunc("/readyz", webHandlers.Readyz)
	r.HandleFunc("/version", Version)
	r.PathPrefix("/").HandlerFunc(webHandlers.Handler)
	log.Info().Msgf("Running HTTP server on %s", addr)
	err := http.ListenAndServe(addr, r)
	log.Fatal().Err(err).Msg("Failed to run HTTP server")
}
