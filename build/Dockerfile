##################################################
# Dependency management stage
FROM golang:1.14.0-alpine3.11 as deps

RUN apk add --no-cache \
    build-base \
    bash

ENV GOPATH=/go
# Enable modules support
ENV GO111MODULE=on

# Install golint
RUN go get golang.org/x/lint/golint

RUN mkdir -p /go/src/github.com/mfamador/pistache/
RUN mkdir -p /go/bin
WORKDIR /go/src/github.com/mfamador/pistache/

# Copy dependencies list to take advantage of Docker cache
COPY go.mod /go/src/github.com/mfamador/pistache/
COPY go.sum /go/src/github.com/mfamador/pistache/

# Install dependencies
RUN go mod download

COPY . /go/src/github.com/mfamador/pistache

##################################################
# Test Chamber 00
FROM deps as tester

# Run linting, formatting checks and run tests
RUN golint -set_exit_status ./... \
    # gofmt always exits with 0, so we have to ask nicely
    && (gofmt -l . | tee /tmp/gofmt.out) && test ! -s /tmp/gofmt.out \
    && go test ./...

##################################################
# Construction work happens here
FROM deps as builder

ENV CGO_ENABLED=0
# Compile static binary
RUN go test ./... \
    && go build -a -o /go/bin/pistache -ldflags "-s -w" github.com/mfamador/pistache/cmd/pistache

##################################################
# Final image here
FROM scratch

WORKDIR /mfamador/app

ENV CONFIGOR_DIR /mfamador/app/configs
ENV CONFIGOR_ENV production

# Copy default config
COPY configs/config.yaml /mfamador/app/configs/config.yaml
# Copy the binary from the previous stage
COPY --from=builder /go/bin/pistache /mfamador/app/pistache

CMD ["/mfamador/app/pistache"]

# Expose the proxy port.
EXPOSE 8080
