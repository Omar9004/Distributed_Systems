# Stage 1: Build
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the entire src folder into the container
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

COPY . .

# Build
RUN go build -race ./main/mrcoordinator.go

FROM alpine:latest
WORKDIR /app/
RUN apk add --no-cache libc6-compat ca-certificates tzdata

#Add the necessary files to run the coordinator
COPY --from=builder /app/mrcoordinator /app/main/pg*txt ./

# Run
CMD ["./mrcoordinator ./pg*txt"]



