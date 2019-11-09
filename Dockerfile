# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:1.13-buster as builder

# Add Maintainer Info
LABEL maintainer="Jean Ribes <jean@ribes.ovh>"

RUN apt-get update && apt-get install go-dep

# Set the Current Working Directory inside the container
#WORKDIR /go/tcp-proxy/src
WORKDIR /go/src/github.com/jeanribes/tcp-reverse-proxy
COPY Gopkg.toml Gopkg.lock ./
# Copy the source from the current directory to the Working Directory inside the container
RUN GOPATH=/go dep ensure --vendor-only

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go
# Build the Go app


######## Start a new stage from scratch #######
#FROM alpine:latest
#
#RUN apk --no-cache add ca-certificates
#
#WORKDIR /root/
FROM busybox
# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/src/github.com/jeanribes/tcp-reverse-proxy/main .

# for server to connect
EXPOSE 1998
# for client to connect
EXPOSE 1999
# for Prometheus metrics
EXPOSE 2112

# Command to run the executable
CMD ["/main"]
