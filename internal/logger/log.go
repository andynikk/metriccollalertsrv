package logger

import (
	"github.com/rs/zerolog"
)

type Logger struct {
	Log zerolog.Logger
}

func (l *Logger) ErrorLog(err error) {
	l.Log.Error().Err(err).Msg("")
}

func (l *Logger) InfoLog(infoString string) {
	l.Log.Info().Msgf("%+v\n", infoString)
}
