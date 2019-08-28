package health_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jsteenb2/health/internal/health"
)

func TestHTTPServer(t *testing.T) {
	t.Run("create new endpoint check", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			svc := &fakeSVC{
				createFn: func(ctx context.Context, endpoint string) (health.Check, error) {
					return health.Check{
						ID:       "id",
						Endpoint: endpoint,
					}, nil
				},
			}

			svr := health.NewHTTPServer(svc)

			body := struct {
				Endpoint string `json:"endpoint"`
			}{
				Endpoint: "https://www.example.com",
			}
			req := httptest.NewRequest(http.MethodPost, "/health/checks", encodeBody(t, body))
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			mustEqual(t, http.StatusCreated, rec.Code, "bad status code")

			var m map[string]string
			decodeBody(t, rec.Body, &m)

			equal(t, "id", m["id"], "invalid id")
			equal(t, body.Endpoint, m["endpoint"], "invalid endpoint")
		})
	})
}

func mustEqual(t *testing.T, expected, got interface{}, msg string) {
	t.Helper()

	if expected != got {
		t.Fatalf("%s: expected=%#v got=%#v", msg, expected, got)
	}
}

func equal(t *testing.T, expected, got interface{}, msg string) {
	t.Helper()

	if expected != got {
		t.Errorf("%s: expected=%#v got=%#v", msg, expected, got)
	}
}

func decodeBody(t *testing.T, r io.Reader, v interface{}) {
	t.Helper()

	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatal(err)
	}
}

func encodeBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		t.Fatal(err)
	}
	return &buf
}

type fakeSVC struct {
	createFn func(ctx context.Context, endpoint string) (health.Check, error)
}

func (f *fakeSVC) Create(ctx context.Context, endpoint string) (health.Check, error) {
	if f.createFn == nil {
		panic("create not implemented")
	}
	return f.createFn(ctx, endpoint)
}
