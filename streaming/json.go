package streaming

import (
	"encoding/json"
	"github.com/boredomist/mixport/mixpanel"
	"io"
	"log"
	"os"
)

// JSONStreamer writes records to a local file in JSON format line by line,
// simply serializing the JSON directly.
//
// Format is simply: `{"key": "value", ...}`. `value` should never be a map,
// but may be a vector. Most values are scalar.
func JSONStreamer(name string, records <-chan mixpanel.EventData) {
	fp, err := os.Create(name)
	if err != nil {
		log.Fatalf("Couldn't create file: %s", err)
	}

	defer func() {
		if err := fp.Close(); err != nil {
			panic(err)
		}
	}()

	encoder := json.NewEncoder(io.Writer(fp))

	for record := range records {
		encoder.Encode(record)
	}
}