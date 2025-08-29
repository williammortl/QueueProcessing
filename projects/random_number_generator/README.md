# Random Number Kafka Producer

Writes random numbers (1..10,000) to a Kafka topic using multiple threads and exposes Prometheus metrics.

## Build

```
make build
```

## Run

```
make run -- -brokers localhost:9092 -topic my_topic -threads 4 -pause 2 -metrics :2112
```

If required parameters are missing, the app prints usage.

Metrics are available at http://localhost:2112/metrics by default.

Flags:
- -brokers: Comma-separated Kafka brokers (required)
- -topic: Kafka topic (required)
- -threads: Number of producer threads (default: CPU count)
- -pause: Pause per thread between writes in seconds (default: 1)
- -metrics: Address to serve Prometheus metrics (default: :2112)
- -kconfig: Opaque additional Kafka config (logged only)

## Docker

Build the container image:

```zsh
make dockerize
```

Run the container (env → flags mapping):

- THREADS → -threads
- PAUSE → -pause
- BROKER → -brokers
- METRICS_PORT → -metrics :PORT

Examples:

1) Demo mode printing numbers to stdout with custom threads/pause and metrics port

```zsh
docker run --rm \
	-e THREADS=4 \
	-e PAUSE=2 \
	-e METRICS_PORT=2112 \
	-p 2112:2112 \
	random-number-generator:latest \
	-demo
```

2) Produce to Kafka (BROKER via env, topic via args)

```zsh
docker run --rm \
	-e THREADS=8 \
	-e PAUSE=1 \
	-e BROKER=localhost:9092 \
	-e METRICS_PORT=2112 \
	-p 2112:2112 \
	random-number-generator:latest \
	-topic my_topic
```

Notes:
- The image entrypoint maps env vars to flags and forwards any additional arguments, so you can pass flags like `-topic`, `-demo`, or `-kconfig` directly after the image name.
- Metrics are exposed on `:$METRICS_PORT` inside the container; publish that port with `-p` to access from the host.
