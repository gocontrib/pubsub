package main

import (
	"fmt"

	"net/http"
	"os"
	"os/signal"

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

func main() {
	addr := opt("PUBSUBD_ADDR", ":4302")
	nats := opt("NATS_URI", "nats:4222")

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

func Logger(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}
