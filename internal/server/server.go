package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	svr *http.Server
}

func New(addr string, h http.Handler) *Server {
	return &Server{
		svr: &http.Server{
			Addr:    addr,
			Handler: h,
		},
	}
}

func (s *Server) Listen(sslEnabled bool, cert, key string) error {
	checkErr := func(err error) error {
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}

	if sslEnabled {
		return checkErr(s.svr.ListenAndServeTLS(cert, key))
	}
	return checkErr(s.svr.ListenAndServe())
}

func (s *Server) Stop(gracePeriod time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()
	return s.svr.Shutdown(ctx)
}
