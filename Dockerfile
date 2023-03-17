# Use the official Golang image as the base image
FROM golang:1.17 as builder

# Set the working directory
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download the Go dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o inazuma .

# Use the scratch image for the final image
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /app/inazuma /inazuma

# Set the entrypoint
ENTRYPOINT ["/inazuma", "serve"]
