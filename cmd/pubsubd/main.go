package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gocontrib/pubsub"
	"github.com/gocontrib/pubsub/sse"
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

	start := func() {
		pubsub.Init(pubsub.HubConfig{
			"url": nats,
		})
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
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Group(func(r chi.Router) {
		// TODO configurable api path
		r.Get("/api/event/stream", sse.GetEventStream)
		r.Get("/api/event/stream/{channel}", sse.GetEventStream)
	})

	return mux
}
