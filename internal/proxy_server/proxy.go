package proxy_server

import (
	"github.com/peterhellberg/link"
	"github.com/terrycain/aws_ecr_proxy/internal/ecr_token"
	"io"
	"net/http"
	"net/url"
	"github.com/rs/zerolog/log"
)

type WebData struct {
	fetcher *ecr_token.EcrFetcher
}

var HttpClient = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (d *WebData) Handler(w http.ResponseWriter, r *http.Request) {
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
	resp, err := HttpClient.Do(req)
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
	_, _ = io.Copy(w, resp.Body)
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
	dst.Header.Add("Authorization", "Basic " + d.fetcher.Token)
}

func FixLinkHeader(scheme, host, header string) (string, error) {
	// Link: <https://000000000000.dkr.ecr.eu-west-2.amazonaws.com/v2/test/tags/list?last=N8jylEwUlHaW6oTKiejfZD%2FAsrOJ0PVfMZA2Me%2F29xo93MTx2AYZCQtZKvYC%2BOMSXdFGYNbSIuuXVqd3ZTlRJfss8wqCyEXe%2BAoLor8QJib0ttd%2BR8t5Trj7R4gx7jtI6d04rCGjY9xIjWpZd9kSsT9iHmFJGYEl8rSHI4a9H%2Bp1MB%2F5leB4%2BQIkYHqidMnmJwZmFn3eZic46QvDdvdM9jXb3TkJFLUQbGxK9sMOHoJI1ELshF3LJiEidUR6SrAWDIfFfOUZlr1iWdS0EA5eGt9NNw9QKxXoXoq6afLMkeUdA8hP40MpiD7%2BUv1pr33fMgDivdTRztZvFUAyPC1%2Frx5j%2BAMt5LBCeeECAArsDC8%3D>; rel="next"
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
			Scheme: scheme,
			Host: host,
			Path: urlPart.Path,
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

func Run(fetcher *ecr_token.EcrFetcher) {

	webHandlers := WebData{fetcher: fetcher}

	mux := http.NewServeMux()
	mux.HandleFunc("/", webHandlers.Handler)
	err := http.ListenAndServe(":8080", mux)
	log.Fatal().Err(err).Msg("Running HTTP server on :8080")
}