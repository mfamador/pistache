// Package models defines the model entities
package models

import "encoding/json"

// Response defines the response entity to cache
type Response struct {
	StatusCode int                 `json:"statusCode"`
	Header     map[string][]string `json:"header"`
	Body       []byte              `json:"body"`
}

// MarshalBinary implements encoding.MarshalBinary interface to be marshaled by redis
func (s *Response) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements encoding.UnmarshalBinary interface to be unmarshaled by redis
func (s *Response) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
