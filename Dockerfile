FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
# Make is requiered for build.
RUN apk update && apk add --no-cache git make ca-certificates

WORKDIR /go/src/github.com/MontFerret/lab

COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux make compile

# Build the final container. And install
FROM microbox/chromium-headless:75.0.3765.1 as runner

RUN apt-get update && apt-get install -y dumb-init

WORKDIR /app

# Add in certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.c

# Add lab binary
COPY --from=builder /go/src/github.com/MontFerret/lab/bin/lab .

EXPOSE 8080

VOLUME /data

ENTRYPOINT ["dumb-init", "--"]
CMD ["/bin/sh", "-c", "chromium --disable-dev-shm-usage --force-gpu-mem-available-mb --full-memory-crash-report --no-sandbox --disable-setuid-sandbox --disable-gpu --headless --remote-debugging-port=9222 & ./lab --wait http://127.0.0.1:9222/json/version --files=file:///data"]