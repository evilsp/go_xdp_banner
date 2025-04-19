FROM golang:1.24 AS build

# china mirror
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd cmd/agent && go build -o agent

FROM debian:bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    busybox curl jq \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /data

COPY --link --from=build /app/cmd/agent/agent /usr/bin/
COPY --link deploy/dockerfile/agent-entrypoint.sh /data/entrypoint.sh

ENTRYPOINT [ "/data/entrypoint.sh" ]