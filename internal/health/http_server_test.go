package health_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jsteenb2/health/internal/health"
)

func TestHTTPServer(t *testing.T) {
	t.Run("create new endpoint check", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			svc := &fakeSVC{
				createFn: func(endpoint string) (health.Check, error) {
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

	t.Run("list provides list of all endpoints paginated", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			svc := &fakeSVC{
				listFn: func(page int) (int, int, []health.Check) {
					out := make([]health.Check, 0, 10)
					for i := 1; i <= 10; i++ {
						out = append(out, health.Check{
							Checked: int64(page*10 + i),
						})
					}
					return 1000, page, out
				},
			}

			svr := health.NewHTTPServer(svc)

			u := url.URL{Path: "/health/checks"}
			params := u.Query()
			params.Set("page", "1")
			u.RawQuery = params.Encode()

			req := httptest.NewRequest(http.MethodGet, u.String(), nil)
			rec := httptest.NewRecorder()

			svr.ServeHTTP(rec, req)

			mustEqual(t, http.StatusOK, rec.Code, "bad status code")

			var resp struct {
				Items []health.Check `json:"items"`
				Page  int            `json:"page"`
				Total int            `json:"total"`
				Size  int            `json:"size"`
			}
			decodeBody(t, rec.Body, &resp)

			equal(t, 1000, resp.Total, "wrong count")
			equal(t, 1, resp.Page, "incorrect page")
			equal(t, 10, resp.Size, "incorrect size")
			mustEqual(t, 10, len(resp.Items), "incorrect number of health checks")
			for i := 0; i < 10; i++ {
				equal(t, int64(i+10+1), resp.Items[i].Checked, "incorrect item received")
			}
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
	createFn func(endpoint string) (health.Check, error)
	listFn   func(page int) (int, int, []health.Check)
}

func (f *fakeSVC) Create(endpoint string) (health.Check, error) {
	if f.createFn == nil {
		panic("create not implemented")
	}
	return f.createFn(endpoint)
}

func (f *fakeSVC) List(page int) (int, int, []health.Check) {
	if f.listFn == nil {
		panic("list not implemented")
	}
	return f.listFn(page)
}
