FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /tamk ./cmd/tamk

FROM alpine:3.21

RUN apk add --no-cache ca-certificates openjdk21-jdk-headless kotlin aapt2 apksigner zipalign

COPY --from=builder /tamk /usr/local/bin/tamk
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080 8765

ENTRYPOINT ["tamk"]
