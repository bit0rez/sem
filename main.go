package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"sem/internal/positions"
)

func main() {
	logger := log.New()
	logger.SetOutput(os.Stdout)
	// TODO: configure log level
	log.SetLevel(log.DebugLevel)
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// TODO: Config and check drivers
	database, err := sql.Open("sqlite3", "./positions.db")
	if err != nil {
		panic(err)
	}
	if err = database.PingContext(context.Background()); err != nil {
		panic(err)
	}
	defer database.Close()

	router := http.NewServeMux()
	server := &http.Server{
		Addr:           ":9080",
		Handler:        router,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 4096,
	}

	// Handle metrics
	router.Handle("/metrics", promhttp.Handler())

	// Subscribe profiler routes
	// TODO: if debug enabled
	router.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	router.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	router.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	router.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// Instantiate service and register service routes
	_, err = positions.NewService(database, router, logger)
	if err != nil {
		logger.Fatalln(err)
	}

	// Register OS signal router
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill, os.Interrupt)
	go func(ch <-chan os.Signal, server *http.Server) {
		<-ch
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		_ = server.Shutdown(ctx)
	}(sigChan, server)

	// Serve
	logger.Infoln("Server started at http://0.0.0.0:9080/")
	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}
	logger.Infoln("Server stopped.")
}
