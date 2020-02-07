package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sem/internal/positions"
)

func main() {
	database, err := sql.Open("sqlite3", "./positions.db")
	if err != nil {
		panic(err)
	}
	if err = database.PingContext(context.Background()); err != nil {
		panic(err)
	}
	defer database.Close()

	logger := log.New(os.Stdout, "[SEM] ", log.LstdFlags)

	p := positions.NewPositions(database, logger)
	err = p.Prepare()
	if err != nil {
		panic(err)
	}

	handler := http.NewServeMux()
	server := &http.Server{
		Addr:           ":9080",
		Handler:        handler,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 4096,
	}

	// Subscribe monitoring route
	handler.Handle("/metrics", promhttp.Handler())

	// Subscribe profiler routes
	handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	handler.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	handler.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	handler.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	handler.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	handler.Handle("/api/summary", p.RequireDomainHandler(http.HandlerFunc(p.HandleSummary)))
	handler.Handle("/api/positions", p.RequireDomainHandler(http.HandlerFunc(p.HandlePositions)))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill, os.Interrupt)
	go func(ch <-chan os.Signal, server *http.Server) {
		<-ch
		_ = server.Close()
	}(sigChan, server)

	logger.Println("Server started at http://0.0.0.0:9080/")
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
	logger.Println("Server stopped.")
}
