package log

import (
	"context"

	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"github.com/sirupsen/logrus"
)

// loggerKey is the context key for the logger.
type loggerKeyType string

var (
	loggerKey loggerKeyType = "logger"
)

// WithLogger returns a new context with the given logger attached to it
func WithLogger(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext returns a logrus.FieldLogger from the given context with the default namespace tags attached to it
func LoggerWithFieldsFromContext(ctx context.Context) logrus.FieldLogger {
	return LoggerWithFieldsFromNamespaceContext(ctx, tag.DefaultNamespace)
}

// LoggerFromContext returns a logrus.FieldLogger from the given context.
// It loads the tags from the provided tags namespace and adds them to the logger.
func LoggerWithFieldsFromNamespaceContext(ctx context.Context, namespaces ...string) logrus.FieldLogger {
	fields := logrus.Fields{}

	if len(namespaces) == 0 {
		namespaces = []string{tag.DefaultNamespace}
	}

	for _, namespace := range namespaces {
		set := tag.FromNamespaceContext(ctx, namespace)
		for _, tag := range set {
			fields[string(tag.Key)] = tag.Value.Interface
		}
	}

	return LoggerFromContext(ctx).WithFields(fields)
}

// LoggerFromContext returns the logrus.FieldLogger attached to given context.
func LoggerFromContext(ctx context.Context) logrus.FieldLogger {
	if logger, ok := ctx.Value(loggerKey).(logrus.FieldLogger); ok {
		return logger
	}
	return logrus.StandardLogger()
}
