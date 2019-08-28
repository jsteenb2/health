package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jsteenb2/health/internal/health"
	"github.com/jsteenb2/health/internal/httpmw"
	"github.com/jsteenb2/health/internal/server"
)

func main() {
	var (
		bindAddr   = flag.String("bind", "127.0.0.1:8080", "address http server listens on")
		sslEnabled = flag.Bool("ssl", false, "enable ssl")
		sslCert    = flag.String("sslcert", "", "ssl certification path")
		sslKey     = flag.String("sslkey", "", "ssl key path")

		filePath      = flag.String("repopath", "endpoints.gob", "file path to the persist the endpoints to disk")
		nukeEndpoints = flag.Bool("nuke", false, "nuke the existing endpoint checks")
	)
	flag.Parse()

	if *nukeEndpoints {
		if err := os.Remove(*filePath); err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}

	healthFileRepo, err := health.NewFileRepository(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	healthSVC := health.NewSVC(healthFileRepo)

	var api http.Handler
	{
		// prefix the health handler with /api and use the behavior of http.StripPrefix
		// to provide a 404 if the route does not have a prefix of /api
		api = http.StripPrefix("/api", health.NewHTTPServer(healthSVC))
		api = httpmw.Recover()(api)
		api = httpmw.ContentType("application/json")(api)
	}

	svr := server.New(*bindAddr, api)
	log.Println("listening at: ", *bindAddr)
	go func(sslEnabled bool, cert, key string) {
		if err := svr.Listen(sslEnabled, cert, key); err != nil {
			log.Println(err)
		}
	}(*sslEnabled, *sslCert, *sslKey)

	<-systemCtx().Done()

	if err := svr.Stop(10 * time.Second); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}

func systemCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopChan
		cancel()
	}()
	return ctx
}
