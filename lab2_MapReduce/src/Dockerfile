# Stage 1: Build
FROM golang:1.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the entire src folder into the container
COPY . .

# Download dependencies
RUN go mod download

# Build the Go application
RUN go build -o app .

# Stage 2: Run
FROM debian:bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the built application from the builder stage
COPY --from=builder /app/app .

# Expose the application port
EXPOSE 1234

# Command to run the application
CMD ["./app"]

