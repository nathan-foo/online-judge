FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o exec-agent ./cmd/exec-agent

FROM golang:1.26.1

RUN apt-get update \
    && apt-get install -y --no-install-recommends util-linux \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -u 10001 -M sandbox

COPY --from=builder /app/exec-agent /usr/local/bin/exec-agent

ENV HOME=/tmp \
    PORT=8000 \
    GOCACHE=/tmp/.cache/go-build \
    GOPATH=/tmp/go \
    GO111MODULE=off \
    GOTOOLCHAIN=local \
    CGO_ENABLED=0 \
    GOMAXPROCS=1
EXPOSE 8000
USER 10001

ENTRYPOINT ["exec-agent"]
