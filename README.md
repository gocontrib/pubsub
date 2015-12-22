[![Build Status](https://travis-ci.org/gocontrib/pubsub.svg?branch=master)](https://travis-ci.org/gocontrib/pubsub)

# pubsub
Simple PubSub interface for golang apps with plugable drivers

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

## TODO
* [x] Continuous integration.
* [ ] Unit tests
* [ ] Example for [Server-sent events](https://en.wikipedia.org/wiki/Server-sent_events)
* [ ] Configuration agnostic removing dependency on "github.com/drone/config" module
* [ ] Get working nsq driver
