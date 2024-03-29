package server

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type StdoutLevelWriter struct {
	io.Writer
}

func (w StdoutLevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level <= zerolog.InfoLevel || level == zerolog.NoLevel {
		return w.Write(p)
	}
	return len(p), nil
}

type StderrLevelWriter struct {
	io.Writer
}

func (w StderrLevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= zerolog.WarnLevel && level < zerolog.NoLevel {
		return w.Write(p)
	}
	return len(p), nil
}

type NoLevelAsNotice struct{}

func (n NoLevelAsNotice) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.NoLevel {
		e.Str("level", "notice")
	}
}

func (s *Server) InitLogger() {
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999-07:00"
	zerolog.SetGlobalLevel(s.Config.Log.GetLevel())

	stdoutWriter := StdoutLevelWriter{Writer: os.Stdout}
	stderrWriter := StderrLevelWriter{Writer: os.Stderr}

	multiWriter := zerolog.MultiLevelWriter(
		&stdoutWriter,
		&stderrWriter,
	)

	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger().Hook(NoLevelAsNotice{})
	log.Log().Msgf("Logger initialized with %s level", zerolog.GlobalLevel())
}
