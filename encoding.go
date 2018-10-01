package pubsub

import (
	"encoding/json"

	"github.com/gocontrib/log"
)

// Marshal message to byte array.
func Marshal(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// Unmarshal message from byte array.
func Unmarshal(data []byte) (interface{}, error) {
	var msg map[string]interface{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Errorf("json.Unmarshal failed: %+v", err)
		return nil, err
	}
	return msg, nil
}
