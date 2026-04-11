// based on https://github.com/tus/tusd/blob/0.10.0/cmd/tusd/cli/hooks.go

package logging

import (
	"encoding/json"
	"strings"

	"github.com/kiwiirc/plugin-fileuploader/events"
	"github.com/rs/zerolog"
)

// tusdLogWriter bridges tusd's standard log.Logger to zerolog.
type tusdLogWriter struct {
	log *zerolog.Logger
}

// TusdLogWriter returns an io.Writer that forwards tusd log output to zerolog at debug level.
func TusdLogWriter(log *zerolog.Logger) *tusdLogWriter {
	return &tusdLogWriter{log: log}
}

func (w *tusdLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	w.log.Debug().Str("source", "tusd").Msg(msg)
	return len(p), nil
}

func TusdLogger(log *zerolog.Logger, broadcaster *events.TusEventBroadcaster) {
	channel := broadcaster.Listen()
	for {
		event, ok := <-channel
		if !ok {
			return // channel closed
		}
		go handleTusEvent(log, event)
	}
}

func handleTusEvent(log *zerolog.Logger, event *events.TusEvent) {
	logEvent := log.Info().
		Str("event", strings.Replace(string(event.Type), "-", "_", -1)).
		Str("id", event.Info.ID).
		Int64("size", event.Info.Size).
		Int64("offset", event.Info.Offset)

	metaCopy := make(map[string]string, len(event.Info.MetaData))
	for k, v := range event.Info.MetaData {
		metaCopy[k] = v
	}
	metadataJSON, err := json.Marshal(metaCopy)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize metadata")
	}
	logEvent.RawJSON("metadata", metadataJSON)

	logEvent.
		Bool("isPartial", event.Info.IsPartial).
		Strs("partialUploads", event.Info.PartialUploads)

	logEvent.Msg("Tusd " + string(event.Type))
}
