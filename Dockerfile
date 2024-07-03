# Use the official Golang image to create a build artifact.
FROM golang:1.21.6

LABEL maintainer="Paul Martin Korp(locopaulito) <paul.m.korp@gmail.com>, Sixten Tedremets"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o llforum ./cmd/server

# Environment variable for database path
ENV FORUM_DB_PATH=/data/forum.db

# Create directory for database file
RUN mkdir -p /data

# Expose port 8080 to the outside world
EXPOSE 8080

# Use the startup script to initiate the application
CMD ["./llforum"]
