package jsonutil

import (
	"bytes"
	"encoding/json"
)

func Stringify(data []byte) string {
	var buf bytes.Buffer

	_ = json.Indent(&buf, data, "", "  ")
	return buf.String()
}
