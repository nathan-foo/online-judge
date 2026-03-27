FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/api-gateway

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/api-gateway ./api-gateway

EXPOSE 8080

USER nonroot:nonroot

CMD ["./api-gateway"]
