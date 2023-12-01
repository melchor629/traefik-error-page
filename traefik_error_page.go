package traefik_error_page

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/melchor629/traefik-error-page/helpers"
)

// Config the plugin configuration.
type Config struct {
	Status    []string `json:"status,omitempty"`
	Service   string   `json:"service,omitempty"`
	Query     string   `json:"query,omitempty"`
	EmptyOnly bool     `json:"emptyOnly,omitempty"`
	Debug     bool     `json:"debug,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Status:    make([]string, 0),
		Service:   "",
		Query:     "/{StatusCode}.html",
		EmptyOnly: true,
		Debug:     false,
	}
}

// The error page plugin.
type ErrorPage struct {
	next             http.Handler
	httpStatusRanges helpers.HTTPCodeRanges
	service          string
	query            string
	name             string
	emptyOnly        bool
	debug            bool
}

// Newly created a Error Page plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Status) == 0 {
		return nil, fmt.Errorf("status cannot be empty")
	}

	if len(config.Service) == 0 {
		return nil, fmt.Errorf("service cannot be empty")
	}

	httpStatusRanges, err := helpers.NewHTTPCodeRanges(config.Status)
	if err != nil {
		return nil, err
	}

	return &ErrorPage{
		httpStatusRanges: httpStatusRanges,
		service:          config.Service,
		query:            config.Query,
		emptyOnly:        config.EmptyOnly,
		debug:            config.Debug,
		next:             next,
		name:             name,
	}, nil
}

func (ep *ErrorPage) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ep.log("request incoming")
	catcher := helpers.NewCodeCatcher(rw, ep.httpStatusRanges, ep.emptyOnly)
	ep.next.ServeHTTP(catcher, req)
	ep.log(fmt.Sprintf("request served, response has filtered code %t and body %t", catcher.IsFilteredCode(), catcher.HasBody()))
	if !catcher.IsFilteredCode() || catcher.HasBody() {
		ep.log("request is OK, should not be handled")
		return
	}

	// check the recorder code against the configured http status code ranges
	code := catcher.GetCode()
	query := ep.parseQuery(code, req.URL.String())
	ep.log(fmt.Sprintf("code is %d and query generated is %s", code, query))

	pageReq, err := newRequest(ep.service + query)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	copyHeaders(pageReq.Header, req.Header)

	ep.log("preparing response modifier from the service response")
	var cm http.ResponseWriter = helpers.NewCodeModifier(rw, code)
	ep.log("preparing request to send to service")
	serviceRequest := pageReq.WithContext(req.Context())
	ep.handleInService(cm, serviceRequest)
}

func (ep *ErrorPage) parseQuery(code int, request_url string) string {
	query := "/" + strings.TrimPrefix(ep.query, "/")
	query = strings.ReplaceAll(query, "{status}", strconv.Itoa(code))
	query = strings.ReplaceAll(query, "{url}", url.QueryEscape(request_url))
	return query
}

func newRequest(baseURL string) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error pages: error when create query: %w", err)
	}

	return req, nil
}

func (ep *ErrorPage) handleInService(rw http.ResponseWriter, req *http.Request) {
	ep.log("making request to service")
	res, err := http.DefaultClient.Do(req)
	ep.log("request made to service")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	if _, err := io.Copy(rw, res.Body); err != nil {
		// TODO handle error
		return
	}

	ep.log("request done to service")
}

func (ep *ErrorPage) log(message string) {
	if ep.debug {
		os.Stdout.WriteString("plugin=traefik-error-page message=\"" + message + "\"\n")
	}
}

func copyHeaders(dst http.Header, src http.Header) {
	for k, vv := range src {
		dst[k] = append(dst[k], vv...)
	}
}
