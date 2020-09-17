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
FROM montferret/chromium:86.0.4240.0 as runner

RUN apt-get update && apt-get install -y dumb-init

# Add in certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.c

# Add lab binary
COPY --from=builder /go/src/github.com/MontFerret/lab/bin/lab .

EXPOSE 8080

ENTRYPOINT ["dumb-init", "--"]
CMD ["/bin/sh", "-c", "./entrypoint.sh & ./worker"]