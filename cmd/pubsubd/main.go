package main

import (
	"fmt"
	"time"

	"net/http"
	"net/url"
	"os"
	"os/signal"

	health "github.com/InVisionApp/go-health/v2"
	"github.com/InVisionApp/go-health/v2/checkers"
	healthHandlers "github.com/InVisionApp/go-health/v2/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gocontrib/pubsub"
	_ "github.com/gocontrib/pubsub/nats"
	_ "github.com/gocontrib/pubsub/redis"
	"github.com/gocontrib/pubsub/sse"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func opt(name, defval string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		return defval
	}
	return val
}

var (
	nats string = opt("NATS_URI", "nats:4222")
)

func main() {
	addr := opt("PUBSUBD_ADDR", ":4302")

	fmt.Printf("starting pubsub --addr %s --nats %s", addr, nats)

	start := func() {
		initHub(nats)
		startServer(addr)
	}

	stop := func() {
		pubsub.Cleanup()
		stopServer()
	}

	die := make(chan bool)
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		<-sig
		die <- true
	}()

	go start()
	<-die

	stop()
}

func initHub(nats string) {
	err := pubsub.Init(pubsub.HubConfig{
		"driver": "nats",
		"url":    nats,
	})
	if err != nil {
		log.Fatalf("cannot initialize hub")
	}
}

var server *http.Server

func startServer(addr string) {
	fmt.Printf("listening %s\n", addr)

	server = &http.Server{
		Addr:    addr,
		Handler: makeHandler(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func stopServer() {
	fmt.Println("shutting down")
	server.Shutdown(nil)
}

func makeHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(Logger)
	r.Use(middleware.Recoverer)

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	r.Group(eventAPI)

	return r
}

func eventAPI(r chi.Router) {
	// TODO configurable api path
	r.Get("/api/event/stream", sse.GetEventStream)
	r.Get("/api/event/stream/{channel}", sse.GetEventStream)
}

func healthAPI(r chi.Router) {
	h := health.New()

	natsURL, err := url.Parse(nats)
	if err != nil {
		log.Fatalf("invalid NATS URL: %v", err)
	}

	nats, err := checkers.NewReachableChecker(&checkers.ReachableConfig{
		URL: natsURL,
	})
	if err != nil {
		log.Fatalf("NewReachableChecker fail for nats: %v", err)
	}

	inerval := time.Duration(10) * time.Second

	h.AddChecks([]*health.Config{
		{
			Name:     "nats",
			Checker:  nats,
			Interval: inerval,
			Fatal:    true,
		},
	})

	if err := h.Start(); err != nil {
		log.Fatalf("unable to start healthcheck: %v", err)
	}

	r.Head("/api/pubsub/health", healthHandlers.NewBasicHandlerFunc(h))
	r.Get("/api/pubsub/health", healthHandlers.NewJSONHandlerFunc(h, nil))
}

func Logger(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}
