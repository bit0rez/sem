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
	"sem/internal/config"
	"sem/internal/positions"
)

func main() {
	// Parse config
	conf, err := config.ParseFlags()
	if err != nil {
		panic(err)
	}

	// Configure logger
	logger := log.New()
	logger.SetOutput(os.Stdout)
	log.SetLevel(log.Level(conf.LogLevel))
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Configure database
	database, err := sql.Open(conf.DbDriver, conf.DbPath)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	// Instantiate HTTP server and router
	router := http.NewServeMux()
	server := &http.Server{
		Addr:           ":9080",
		Handler:        router,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 512,
	}

	// Handle metrics
	router.Handle("/metrics", promhttp.Handler())

	// Subscribe profiler routes
	if conf.DebugMode {
		router.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		router.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		router.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		router.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}

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
