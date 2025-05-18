set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

go build -o build/xdp-server ./cmd/orch
go build -o build/xdp-agent ./cmd/agent