# Production stage
FROM golang:1.24-alpine AS builder

# Move to working directory (/build).
WORKDIR /build

# Download ca certificates
RUN apk add --no-cache ca-certificates

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container.
COPY . .

# Set necessary environment variables needed for our image and build the API server.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o apiserver cmd/server/main.go

# Compress the binary
# RUN upx --best --lzma coreServer

FROM scratch AS production

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/build/apiserver", "/"]
COPY --from=builder ["/build/certs/prod", "/certs/prod"]

# Copy the CA certificates from the builder stage.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Command to run when starting the container.
ENTRYPOINT ["/apiserver"]