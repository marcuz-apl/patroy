# Stage 1: Build static Go binary
FROM golang:alpine AS builder

WORKDIR /src

# Install git and make for Alfazen build metadata
RUN apk add --no-cache git make

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .

# Compile static binary with Alfazen versioning flags
RUN make build

# Stage 2: Ultra-lean production container
FROM alpine:3.20

# Install Chromium, fonts, and runtime dependencies
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    tzdata

# Create unprivileged application user
RUN addgroup -S patroy && adduser -S -G patroy patroy

# Copy compiled binary from builder
COPY --from=builder /src/bin/patroy /usr/local/bin/patroy

# Ensure Chromium discovery defaults to system package
ENV PATROY_CHROME_BIN=/usr/bin/chromium-browser

USER patroy
WORKDIR /home/patroy

EXPOSE 4023

ENTRYPOINT ["/usr/local/bin/patroy"]
CMD ["serve", "--host", "0.0.0.0", "--port", "4023"]
