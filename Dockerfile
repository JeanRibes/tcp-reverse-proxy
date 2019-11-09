# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Jean Ribes <jean@ribes.ovh>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go


######## Start a new stage from scratch #######
#FROM alpine:latest
#
#RUN apk --no-cache add ca-certificates
#
#WORKDIR /root/
FROM busybox
# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

EXPOSE 1998 # for server to connect
EXPOSE 1999 # for client to connect
EXPOSE 2112 # for Prometheus metrics

# Command to run the executable
CMD ["/main"]
