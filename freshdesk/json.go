package freshdesk

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// structToMap renders v through its JSON tags into a mutable map. Request types
// use it to express semantics struct tags cannot: sending an explicit null to
// clear a field, or an empty array to clear a collection, while still omitting
// every field the caller did not set.
func structToMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}
	m := map[string]any{}
	// UseNumber keeps integer IDs from round-tripping through float64 and
	// re-emerging in scientific notation.
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("decoding request: %w", err)
	}
	return m, nil
}

// marshalMap encodes a request map without HTML-escaping, so signatures and
// article bodies containing markup survive the round trip unchanged.
func marshalMap(m map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
