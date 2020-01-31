############################
# STEP 1 build executable binary
############################
FROM golang:1.13-alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update \
    && apk add --no-cache git ca-certificates tzdata alpine-sdk \
    && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /go/src/github.com/andig/gravo

COPY . .
RUN make build

#############################
## STEP 2 build a small image
#############################
FROM alpine

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /go/src/github.com/andig/gravo/gravo /usr/bin/gravo

# Use an unprivileged user.
USER appuser

EXPOSE 8000

# Run the binary.
ENTRYPOINT ["/usr/bin/gravo"]
