package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func jsonUnmarshalStrict(content []byte, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode JSON: trailing content")
	}
	return nil
}
