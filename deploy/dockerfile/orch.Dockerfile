FROM golang:1.24 AS build

# china mirror
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd cmd/orch && go build -o orch

FROM busybox:1.35.0-uclibc as busybox

# must use cgo because there is no liibc in static
FROM gcr.io/distroless/static-debian12

COPY --link --from=busybox /bin /bin

WORKDIR /data

COPY --link --from=build /app/cmd/orch/orch /bin/

ENTRYPOINT [ "/bin/orch" ]