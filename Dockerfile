# Use an official Golang image as the base image
FROM golang:latest

# Set a default port
ARG SERVER_PORT=2000
ENV PORT=${SERVER_PORT}

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Create a directory for your Go application
WORKDIR /app

# Copy your Go application source code to the container
COPY . /app

# Build the Go application
RUN go build -o wopr wopr.go

# Expose the specified port
EXPOSE ${PORT}

# Define the CMD with optional flags
CMD ["./wopr"]
