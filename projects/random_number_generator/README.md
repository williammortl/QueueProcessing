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
