# Define directories
PROTO_SRC_DIR := .
PROTO_OUT_DIR := $(PROTO_SRC_DIR)

# Protobuf compiler and plugins
PROTOC := protoc
PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := protoc-gen-grpc-gateway

# GRPC Gateway specific variables
GATEWAY_API_CONFIG_FILE_NAME :=  "gateway.yaml"

# Ensure the necessary plugins are installed
.PHONY: check-plugins
check-plugins:
	@which $(PROTOC) > /dev/null || (echo "protoc is not installed"; exit 1)
	@which $(PROTOC_GEN_GO) > /dev/null || (echo "protoc-gen-go is not installed"; exit 1)
	@which $(PROTOC_GEN_GO_GRPC) > /dev/null || (echo "protoc-gen-go-grpc is not installed"; exit 1)
	@which $(PROTOC_GEN_GRPC_GATEWAY) > /dev/null || (echo "protoc-gen-grpc-gateway is not installed"; exit 1)


# Find all .proto files recursively in the source directory
PROTO_FILES := $(shell find $(PROTO_SRC_DIR) -name '*.proto')

# Generate .pb.go files
.PHONY: generate
generate: check-plugins
	@echo "Compiling Protobuf files..."
# Grpc
	@for proto in $(PROTO_FILES); do \
		echo "Processing $$proto..."; \
		$(PROTOC) --proto_path=$(PROTO_SRC_DIR) \
		          --go_out=$(PROTO_OUT_DIR) \
		          --go-grpc_out=$(PROTO_OUT_DIR) \
		          $$proto || exit 1; \
	done
# Gateway
	@for proto in $(PROTO_FILES); do \
        DIRNAME=$$(dirname $$proto); \
        GATEWAY_API_CONFIG_FILE=$$DIRNAME/$(GATEWAY_API_CONFIG_FILE_NAME); \
        if [ -f "$$GATEWAY_API_CONFIG_FILE" ]; then \
            echo "Processing $$proto for gateway..."; \
            $(PROTOC) \
                --proto_path=$(PROTO_SRC_DIR) \
                --grpc-gateway_out=$(PROTO_OUT_DIR) \
                --grpc-gateway_opt paths=source_relative \
                --grpc-gateway_opt grpc_api_configuration=$$GATEWAY_API_CONFIG_FILE \
                $$proto || { echo "Error: Failed to process $$proto"; exit 1; }; \
        else \
            echo "API configuration file $$GATEWAY_API_CONFIG_FILE not found, skipping..."; \
        fi; \
    done

	@echo "Protobuf compilation completed."

# Clean up generated files
.PHONY: clean
clean:
	@echo "Cleaning up generated files..."
	@find $(PROTO_SRC_DIR) -name '*.pb.go' -delete
	@find $(PROTO_SRC_DIR) -name '*.pb.gw.go' -delete
	@echo "Cleanup completed."
