package log

import "go.uber.org/zap/zapcore"

func NewTreeWithDefaultLogger(cores ...zapcore.Core) zapcore.Core {
	def := newDefaultProductionLog()

	cores = append(cores, def.Core())
	return zapcore.NewTee(cores...)
}
