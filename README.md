# Inazuma

Inazuma is a front-end proxy server with cache capability, built using Go. It uses S3-like object storage as the cache database and can store the content of HTTP GET requests in the object storage. It can also accept HTTP PURGE requests to update the cache in a queue, processing the requests asynchronously to avoid excessive load on the upstream server.

## Features

- Caching of HTTP GET requests in S3-like object storage
- Support for cache bypass based on specific cookies or query parameters
- Support for different Chinese-language variants through the `Accept-Language` header
- Asynchronous cache updates using Kafka and Redis
- Docker and GitHub Actions integration for easy deployment

## Installation

### Prerequisites

- Go (1.17 or later)
- An S3-like object storage service (e.g., Amazon S3, MinIO)
- A Kafka server
- A Redis server

### Building the application

1. Clone the repository:

```bash
git clone https://github.com/mudkipme/inazuma.git
```

2. Change into the `inazuma` directory:

```bash
cd inazuma
```

3. Build the Go application:

```bash
go build -o inazuma
```

## Configuration

Configure the application using the following environment variables:

- `UPSTREAM_SERVER`: The upstream server's URL (e.g., `http://example.com`)
- `S3_ENDPOINT`: The S3-compatible object storage endpoint (e.g., `http://localhost:9000`)
- `S3_REGION`: The S3-compatible object storage region (e.g., `us-east-1`)
- `S3_ACCESS_KEY`: The S3-compatible object storage access key
- `S3_SECRET_KEY`: The S3-compatible object storage secret key
- `S3_BUCKET`: The S3-compatible object storage bucket name
- `KAFKA_BROKER`: The Kafka broker's address (e.g., `localhost:9092`)
- `KAFKA_TOPIC`: The Kafka topic for PURGE requests
- `REDIS_ADDR`: The Redis server's address (e.g., `localhost:6379`)
- `REDIS_PASSWORD`: The Redis server's password (if any)
- `REDIS_DB`: The Redis server's database number

## Usage

Run the application with the following command:

```bash
./inazuma
```

The proxy server will start listening on port 8080 by default.

## License

Inazuma is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

The initial commit of this project is 100% created with GPT-4, including the README file. Here are the [screenshots](https://github.com/mudkipme/inazuma/tree/0.0.1/screenshots) of the conversation between me and the AI model.
