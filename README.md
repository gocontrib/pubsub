[![Build Status](https://travis-ci.org/gocontrib/pubsub.svg?branch=master)](https://travis-ci.org/gocontrib/pubsub)

# pubsub
Simple PubSub interface for golang apps with plugable drivers.

## Supported drivers
* in-memory implementation based on go channels
* [nats.io](http://nats.io/)
* redis using [redigo](https://github.com/garyburd/redigo)
* [nsq.io](http://nsq.io/) - draft, not completed!

## API

```go
// Hub interface of pubsub system.
type Hub interface {
	// Publish sends input message to specified channels.
	Publish(channels []string, msg interface{})
	// Subscribe opens channel to listen specified channels.
	Subscribe(channels []string) (Channel, error)
	// Close stops the pubsub hub.
	Close() error
}

// Channel to listen pubsub events.
type Channel interface {
	// Read returns channel to receive events.
	Read() <-chan interface{}
	// Close stops listening underlying pubsub channels.
	Close() error
	// CloseNotify returns channel to receive event when this channel is closed.
	CloseNotify() <-chan bool
}
```

## Server-sent events

Example how to implement SSE using this pubsub package:

```go
package examples

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gocontrib/log"
	"github.com/gocontrib/pubsub"
)

// SendEvents streams events from specified channels as Server Sent Events packets
func SendEvents(w http.ResponseWriter, r *http.Request, channels []string) {
	// make sure that the writer supports flushing
	flusher, ok := w.(http.Flusher)

	if !ok {
		log.Error("current response %T does not implement http.Flusher, plase check your middlewares that wraps response", w)
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		log.Error("current response %T does not implement http.CloseNotifier, plase check your middlewares that wraps response", w)
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	sub, err := pubsub.Subscribe(channels)
	if err != nil {
		log.Error("subscribe failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set the headers related to event streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// channel to send data to client
	var send = make(chan []byte)
	var closeConn = closeNotifier.CloseNotify()
	var heartbeatTicker = time.NewTicker(10 * time.Second) // TODO get from config
	var heatbeat = []byte{}

	var stop = func() {
		log.Info("SSE streaming stopped")
		heartbeatTicker.Stop()
		sub.Close()
		close(send)
	}

	// goroutine to listen all room events
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("recovered from panic: %+v", err)
				debug.PrintStack()
			}
		}()

		log.Info("SSE streaming started")

		for {
			select {
			case msg := <-sub.Read():
				var err error
				var data, ok = msg.([]byte)
				if !ok {
					data, err = json.Marshal(msg)
					if err != nil {
						// TODO should we ignore error messages?
						log.Error("json.Marshal failed with: %+v", err)
						continue
					}
				}
				if len(data) == 0 {
					log.Warning("empty message is not allowed")
					continue
				}
				send <- data
			// listen to connection close and un-register message channel
			case <-sub.CloseNotify():
				log.Info("subscription closed")
				stop()
				return
			case <-closeConn:
				stop()
				return
			case <-heartbeatTicker.C:
				send <- heatbeat
			}
		}
	}()

	for {
		var data, ok = <-send
		if !ok {
			log.Info("connection closed, stop streaming of %v", channels)
			return
		}

		if len(data) == 0 {
			fmt.Fprint(w, ":heartbeat signal\n\n")
		} else {
			fmt.Fprintf(w, "data: %s\n\n", data)
		}

		flusher.Flush()
	}
}
```

## TODO
* [x] continuous integration
* [ ] (in progress) unit tests
* [x] example for [Server-sent events](https://en.wikipedia.org/wiki/Server-sent_events)
* [ ] example for websocket event streams
* [x] remove dependency on "github.com/drone/config" module
* [ ] get working nsq driver
* [ ] find and implement more drivers maybe faster then nats.io
  * [ ] [sarama](https://github.com/Shopify/sarama) client for [Apache Kafka](https://kafka.apache.org/)
  * [ ] [zmq4](https://github.com/pebbe/zmq4) client for [zeromq](http://zeromq.org/)


