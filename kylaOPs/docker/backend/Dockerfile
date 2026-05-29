# Backend image for kylaBE.
# Build context: monorepo root (kylaCX/). See docker-compose.yml.

FROM golang:1.24-alpine AS builder
WORKDIR /build
RUN apk add --no-cache ca-certificates
COPY kylaBE/go.mod kylaBE/go.sum ./
RUN go mod download
COPY kylaBE/ ./
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o /out/apiserver ./cmd/server

FROM alpine:3.20 AS production
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S kyla && adduser -S kyla -G kyla
WORKDIR /app
COPY --from=builder /out/apiserver /app/apiserver
COPY --from=builder /build/certs /app/certs
USER kyla
EXPOSE 50050 8085
ENTRYPOINT ["/app/apiserver"]
