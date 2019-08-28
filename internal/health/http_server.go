package health

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type HTTPServer struct {
	mux *http.ServeMux
	svc SVC
}

func NewHTTPServer(svc SVC) *HTTPServer {
	svr := &HTTPServer{
		svc: svc,
	}

	mux := http.NewServeMux()
	{
		mux.Handle("/health/checks", http.StripPrefix("/health", http.HandlerFunc(svr.checkRoutes)))
	}
	svr.mux = mux

	return svr
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *HTTPServer) checkRoutes(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/checks":
		switch r.Method {
		case http.MethodGet:
			s.list(w, r)
		case http.MethodPost:
			s.create(w, r)
		default:
			http.Error(w, "unsupported HTTP method", http.StatusMethodNotAllowed)
			return
		}
	}
}

func (s *HTTPServer) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := s.svc.Create(body.Endpoint)
	if err != nil {
		switch err {
		case errInvalidEndpoint, errEndpointExists:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			http.Error(w, "unexpected error", http.StatusInternalServerError)
		}
		return
	}

	resp := struct {
		ID       string `json:"id"`
		Endpoint string `json:"endpoint"`
	}{
		ID:       c.ID,
		Endpoint: c.Endpoint,
	}

	w.WriteHeader(http.StatusCreated)
	if err := prettyEncoder(w).Encode(resp); err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) list(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))

	total, currentPage, c := s.svc.List(page)

	w.WriteHeader(http.StatusOK)

	body := struct {
		Items []Check `json:"items"`
		Page  int     `json:"page"`
		Total int     `json:"total"`
		Size  int     `json:"size"`
	}{
		Items: c,
		Page:  currentPage,
		Total: total,
		Size:  10,
	}

	err := prettyEncoder(w).Encode(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func prettyEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc
}
