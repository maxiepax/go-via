package main

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Formatter implements logrus.Formatter interface.
type Formatter struct {
	// Timestamp format
	TimestampFormat string
	// Available standard keys: time, msg, lvl
	// Also can include custom fields but limited to strings.
	// All of fields need to be wrapped inside %% i.e %time% %msg%
	LogFormat string
	Level     logrus.Level
}

func ConfigureLogger() {

	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			Level:           logrus.DebugLevel,
		})
		logrus.SetLevel(logrus.DebugLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

const (
	// Default log format will output -> 2006-01-02 15:04:05 severity=LEVEL msg="Great Success!"
	defaultLogFormat = "%time% severity=%lvl% msg=\"%msg%\""
	// Default debug log format will output -> 2006-01-02 15:04:05 severity=LEVEL main.go:69 msg="Great Success!"
	defaultDebugLogFormat  = "%time% severity=%lvl% %file%:%line% msg=\"%msg%\""
	defaultTimestampFormat = time.RFC3339
)

// Format building log message.
// All levels except Debug: 2006-01-02 15:04:05 severity=info msg="Something happened" [key=value, key=value,..]
// e.g 2006-01-02 15:04:05 severity=info msg="Something happened" customer=ix release=1.0 jobId=123456
// All levels except Debug: 2006-01-02 15:04:05 severity=info msg="Something happened" caller=main.go:69 [key=value, key=value,..]
// e.g 2006-01-02 15:04:05 severity=debug msg="Something happened" caller=main.go:69 customer=ix release=1.0 jobId=123456

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := f.LogFormat
	if output == "" && f.Level != logrus.DebugLevel {
		output = defaultLogFormat
	}

	if output == "" && f.Level == logrus.DebugLevel {
		output = defaultDebugLogFormat
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}

	output = strings.Replace(output, "%time%", entry.Time.Format(timestampFormat), 1)

	output = strings.Replace(output, "%msg%", entry.Message, 1)

	if f.Level == logrus.DebugLevel {
		elements := strings.Split(entry.Caller.File, "/")
		file := elements[len(elements)-1]
		output = strings.Replace(output, "%file%", file, 1)

		output = strings.Replace(output, "%line%", strconv.Itoa(entry.Caller.Line), 1)
	}

	level := strings.ToUpper(entry.Level.String())
	output = strings.Replace(output, "%lvl%", level, 1)

	keys := make([]string, 0, len(entry.Data))

	for k := range entry.Data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		switch v := entry.Data[k].(type) {
		case string:
			field := strings.Join([]string{k, v}, "=")
			output = strings.Join([]string{output, field}, " ")
		case int:
			s := strconv.Itoa(v)
			field := strings.Join([]string{k, s}, "=")
			output = strings.Join([]string{output, field}, " ")
		case bool:
			s := strconv.FormatBool(v)
			field := strings.Join([]string{k, s}, "=")
			output = strings.Join([]string{output, field}, " ")
		}
	}

	return []byte(output + "\n"), nil
}
