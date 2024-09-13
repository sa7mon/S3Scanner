package log

import (
	log "github.com/sirupsen/logrus"
	"io"
)

// Code from: https://github.com/sirupsen/logrus/issues/894#issuecomment-1284051207

// FormatterHook is a hook that writes logs of specified LogLevels with a formatter to specified Writer
type FormatterHook struct {
	Writer        io.Writer
	LogLevels     []log.Level
	Formatter     log.Formatter
	DefaultFields log.Fields
}

// Fire will be called when some logging function is called with current hook
// It will format log entry and write it to appropriate writer
func (hook *FormatterHook) Fire(entry *log.Entry) error {
	// Add default fields to any set by the user (entry.Data)
	newFieldsMap := make(log.Fields, len(hook.DefaultFields)+len(entry.Data))
	for k, v := range entry.Data {
		newFieldsMap[k] = v
	}
	for k, v := range hook.DefaultFields {
		newFieldsMap[k] = v
	}
	entry.Data = newFieldsMap

	line, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(line)
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *FormatterHook) Levels() []log.Level {
	return hook.LogLevels
}
