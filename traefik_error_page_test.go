package traefik_error_page_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/melchor629/traefik-error-page"
)

func TestFailsIfEmptyStatus(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.Service = "https://http.cat"
	cfg.Query = "/{status}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(404)
	})

	_, err := plugin.New(ctx, next, cfg, "demo-plugin")
	if err == nil {
		t.Errorf("expected plugin constructor to fail")
	}

	if err.Error() != "status cannot be empty" {
		t.Errorf("expected error is incorrect: %s", err.Error())
	}
}

func TestFailsIfEmptyService(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.Status = []string{"400-499", "500-599"}
	cfg.Query = "/{status}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(404)
	})

	_, err := plugin.New(ctx, next, cfg, "demo-plugin")
	if err == nil {
		t.Errorf("expected plugin constructor to fail")
	}

	if err.Error() != "service cannot be empty" {
		t.Errorf("expected error is incorrect: %s", err.Error())
	}
}

func TestFailsIfInvalidStatus(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.Status = []string{"e"}
	cfg.Service = "https://http.cat"
	cfg.Query = "/{status}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(404)
	})

	_, err := plugin.New(ctx, next, cfg, "demo-plugin")
	if err == nil {
		t.Errorf("expected plugin constructor to fail")
	}

	if err.Error() != "strconv.Atoi: parsing \"e\": invalid syntax" {
		t.Errorf("expected error is incorrect: %s", err.Error())
	}
}

func TestErrorWithEmptyResponse(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.Status = []string{"400-499", "500-599"}
	cfg.Service = "https://http.cat"
	cfg.Query = "/{status}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(404)
	})

	handler, err := plugin.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Code != 404 {
		t.Errorf("status code is not 404: %d", recorder.Code)
	}

	assertHasServed(t, recorder, true)
	assertHeader(t, recorder, "Content-Type", "")
	assertHeader(t, recorder, "X-Content-Type-Options", "")
}

func TestErrorWithResponse(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.Status = []string{"400-499", "500-599"}
	cfg.Service = "https://http.cat"
	cfg.Query = "/{status}"

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.NotFound(rw, req)
	})

	handler, err := plugin.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Code != 404 {
		t.Errorf("status code is not 404: %d", recorder.Code)
	}

	assertHasServed(t, recorder, false)
	assertHeader(t, recorder, "Content-Type", "text/plain; charset=utf-8")
	assertHeader(t, recorder, "X-Content-Type-Options", "nosniff")
}

func TestErrorWithResponseButForceHandle(t *testing.T) {
	cfg := plugin.CreateConfig()
	cfg.Status = []string{"400-499", "500-599"}
	cfg.Service = "https://http.cat"
	cfg.Query = "/{status}"
	cfg.EmptyOnly = false

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.NotFound(rw, req)
	})

	handler, err := plugin.New(ctx, next, cfg, "demo-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	if recorder.Code != 404 {
		t.Errorf("status code is not 404: %d", recorder.Code)
	}

	assertHasServed(t, recorder, true)
	assertHeader(t, recorder, "Content-Type", "")
	assertHeader(t, recorder, "X-Content-Type-Options", "")
}

func assertHasServed(t *testing.T, recorder *httptest.ResponseRecorder, served bool) {
	t.Helper()

	value := recorder.Header().Get("X-ErrorPage")
	if served && value != "served" {
		t.Errorf("the response has not been served from the error service: %s", value)
	}
	if !served && value != "" {
		t.Errorf("the response has been served from the error service and should not")
	}
}

func assertHeader(t *testing.T, recorder *httptest.ResponseRecorder, key, value string) {
	t.Helper()

	current_value := recorder.Header().Get(key)
	if current_value != value {
		t.Errorf("the header %s does not match expected value %s: %s", key, value, current_value)
	}
}
