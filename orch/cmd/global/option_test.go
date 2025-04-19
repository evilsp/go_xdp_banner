package global

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	v_test, controller_config, err := LoadConfig("../../../configs/controller/values.yaml")
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}
	fmt.Printf("Controller Name: %s\n", v_test.Get("controllerName"))
	fmt.Printf("Controller Otel Endpoint: %s\n", controller_config.Trace.Otel.Endpoint)
}
