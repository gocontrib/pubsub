[![Build Status](https://travis-ci.org/gocontrib/pubsub.svg?branch=master)](https://travis-ci.org/gocontrib/pubsub)

# pubsub

Simple PubSub interface for golang apps with plugable drivers.

## Supported drivers
* in-memory implementation based on go channels
* [nats.io](http://nats.io/)
* redis using [redigo](https://github.com/garyburd/redigo)
* [nsq.io](http://nsq.io/) - draft, not completed!

## API

API is pretty simple

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

See code of [built-in package](https://github.com/gocontrib/pubsub/blob/master/sse/sse.go)

## TODO
* [x] continuous integration
* [ ] (in progress) unit tests
* [ ] example for websocket event streams
* [x] remove dependency on "github.com/drone/config" module
* [ ] get working nsq driver
* [ ] find and implement more drivers maybe faster then nats.io
  * [ ] [sarama](https://github.com/Shopify/sarama) client for [Apache Kafka](https://kafka.apache.org/)
  * [ ] [zmq4](https://github.com/pebbe/zmq4) client for [zeromq](http://zeromq.org/)

