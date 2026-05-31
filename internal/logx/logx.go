// Package logx is gogen's own structured logger. It writes coloured, human
// readable logs to stderr (keeping stdout free for command results). The level
// is taken from the GOGEN_LOG environment variable (debug|info|warn|error).
package logx

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init installs a console zap logger as the global logger and returns it.
func Init() *zap.Logger {
	lvl := zapcore.InfoLevel
	if v := os.Getenv("GOGEN_LOG"); v != "" {
		if parsed, err := zapcore.ParseLevel(strings.ToLower(v)); err == nil {
			lvl = parsed
		}
	}

	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), zapcore.Lock(os.Stderr), lvl)

	l := zap.New(core)
	zap.ReplaceGlobals(l)
	return l
}

// L returns the global structured logger.
func L() *zap.Logger { return zap.L() }

// S returns the global sugared logger.
func S() *zap.SugaredLogger { return zap.S() }
