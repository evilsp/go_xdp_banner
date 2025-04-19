package common

import "encoding/json"

// MustMarshal is a helper function to marshal a value to JSON and panic on error.
func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
