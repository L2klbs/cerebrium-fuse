# Use official Go 1.24.3 base image
FROM golang:1.24.3-bullseye

# Install FUSE
RUN apt-get update && apt-get install -y fuse3 && rm -rf /var/lib/apt/lists/*

# Allow FUSE to run
RUN mkdir -p /mnt/all-projects

# Copy application source code
WORKDIR /app
COPY . .

# Tidy dependencies
RUN go mod tidy

RUN go build -o main .
CMD ["./main"]