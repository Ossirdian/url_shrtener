FROM golang:1.26.0-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /url-shortener ./cmd/url_shortener

FROM alpine:3.22

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /url-shortener /usr/local/bin/url-shortener
COPY --from=builder /src/migrations /app/migrations

USER 65532:65532
EXPOSE 8084
ENTRYPOINT ["url-shortener"]
