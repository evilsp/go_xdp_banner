package otlp

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func Meter() metric.Meter {
	return otel.Meter("orch")
}
