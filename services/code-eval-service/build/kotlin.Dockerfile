FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o exec-agent ./cmd/exec-agent

FROM eclipse-temurin:21-jdk-jammy

ARG KOTLIN_VERSION=2.0.20

RUN apt-get update \
    && apt-get install -y --no-install-recommends util-linux curl unzip \
    && curl -fsSL -o /tmp/kotlin.zip \
        https://github.com/JetBrains/kotlin/releases/download/v${KOTLIN_VERSION}/kotlin-compiler-${KOTLIN_VERSION}.zip \
    && unzip -q /tmp/kotlin.zip -d /opt \
    && rm /tmp/kotlin.zip \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -u 10001 -M sandbox

COPY --from=builder /app/exec-agent /usr/local/bin/exec-agent

ENV HOME=/tmp PORT=8000 PATH="/opt/kotlinc/bin:${PATH}"
EXPOSE 8000
USER 10001

ENTRYPOINT ["exec-agent"]
