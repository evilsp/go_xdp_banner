package otlp

import (
	"context"
	"errors"
	"fmt"
	"os"

	"xdp-banner/pkg/log"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func setOtelEndpointEnv(endpoint string) {
	setEnv("OTEL_EXPORTER_OTLP_ENDPOINT", endpoint)
}

func setEnv(key string, value string) {
	_ = os.Setenv(key, value)
}

func MergeWithDefaultResource(name string, res *resource.Resource) (*resource.Resource, error) {
	defaultRes, err := DefaultResource(name)
	if err != nil {
		return nil, fmt.Errorf("get default resource failed: %w", err)
	}

	return resource.Merge(res, defaultRes)
}

func DefaultResource(name string) (*resource.Resource, error) {
	detectedRes, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),         // Discover and provide host information.
	)
	if err != nil {
		if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
			// non critical error
			log.Warn("otel partial resource detected", log.StringField("error", err.Error()))
		} else {
			return nil, fmt.Errorf("get otel resource failed: %w", err)
		}
	}

	serverRes := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(name),
	)

	res, err := resource.Merge(detectedRes, serverRes)
	if err != nil {
		return nil, fmt.Errorf("otel merge resource failed: %w", err)
	}

	return res, nil
}
