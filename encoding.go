package pubsub

import (
	"bytes"
	"encoding/gob"

	"github.com/gocontrib/log"
)

// Box of any value.
type Box struct {
	Value interface{}
}

// Marshal message to byte array.
func Marshal(value interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(&Box{value})
	if err != nil {
		log.Error("gob.Encode failed: %+v", err)
		return nil, err
	}
	return b.Bytes(), nil
}

// Unmarshal message from byte array.
func Unmarshal(data []byte) (interface{}, error) {
	var box Box
	var b = bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&box)
	if err != nil {
		log.Error("gob.Decode failed: %+v", err)
		return nil, err
	}
	return box.Value, nil
}
