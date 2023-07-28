package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
)

type NestedJSONFormatter struct {
	PrettyPrint       bool
	DisableHTMLEscape bool
}

func (f *NestedJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	data["level"] = entry.Level.String()
	data["msg"] = entry.Message
	data["time"] = entry.Time

	val, ok := entry.Data["bucket"]
	if ok {
		data["bucket"] = val
	}

	b := &bytes.Buffer{}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!f.DisableHTMLEscape)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %w", err)
	}

	return b.Bytes(), nil
}
