package database

import (
	"context"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
	gormLog "gorm.io/gorm/logger"
	"strconv"
	"time"
)

var (
	levelMap = map[gormLog.LogLevel]zerolog.Level{
		gormLog.Silent: zerolog.NoLevel,
		gormLog.Info:   zerolog.InfoLevel,
		gormLog.Warn:   zerolog.WarnLevel,
		gormLog.Error:  zerolog.ErrorLevel,
	}
)

type gormLogger struct {
	log zerolog.Logger
}

func NewGormLogger() gormLog.Interface {
	return &gormLogger{
		log: logger.GetPackageLogger("gorm"),
	}
}

func (gl *gormLogger) LogMode(level gormLog.LogLevel) gormLog.Interface {
	g := *gl
	zerologLevel, ok := levelMap[level]
	if !ok {
		zerologLevel = zerolog.InfoLevel
	}
	g.log = g.log.Level(zerologLevel)
	return &g
}

func (gl *gormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	gl.log.Info().Msgf("%s; data=%#v", msg, data)
}

func (gl *gormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	gl.log.Warn().Msgf("%s; data=%#v", msg, data)
}

func (gl *gormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	gl.log.Error().Msgf("%s; data=%#v", msg, data)
}

func (gl *gormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if gl.log.GetLevel() != zerolog.TraceLevel {
		return
	}
	elapsed := time.Since(begin)
	traceLog := gl.log.With().Str("elapsed", elapsed.String()).Logger()
	sql, rows := fc()
	if err != nil {
		traceLog = traceLog.With().Str("err", err.Error()).Logger()
	}
	if rows > -1 {
		rowsStr := strconv.Itoa(int(rows))
		traceLog = traceLog.With().Str("rows", rowsStr).Logger()
	}
	traceLog.Trace().Msg(sql)
}
