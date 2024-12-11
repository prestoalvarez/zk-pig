package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// WriteJSON writes a JSON response with the given status code and data
func WriteJSON(rw http.ResponseWriter, statusCode int, data interface{}) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	return json.NewEncoder(rw).Encode(data)
}

// ErrorRespMsg is a struct representing an error response message
type ErrorRespMsg struct {
	Message string `json:"message" example:"error message"`
	Code    string `json:"status,omitempty" example:"IR001"`
} // @name Error

// WriteError writes an error response with the given status code and error
func WriteError(rw http.ResponseWriter, statusCode int, err error) {
	_ = WriteJSON(rw, statusCode, ErrorRespMsg{
		Message: err.Error(),
	})
}

// DecodeJSON decodes a JSON request body into the given object
func DecodeJSON(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return json.NewDecoder(req.Body).Decode(obj)
}

// UnmarshalQuery unmarshals a URL query into the given object
func UnmarshalQuery(params url.Values, obj interface{}) error {
	m := map[string]string{}
	for k, v := range params {
		m[k] = v[0]
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}

	return nil
}
