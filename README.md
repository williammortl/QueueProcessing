# QueueProcessing
An example of a service writing to a Kafka Queue with a Storm Cluster w/ spout and bolt processing it

## Random Number Kafka Producer

This repo includes a Go service that writes random numbers (1..10,000) to a Kafka topic using multiple threads and exposes Prometheus metrics.

### Build

Make sure you have Go 1.22+ installed.

```
make build
```

### Run

```
make run -- -brokers localhost:9092 -topic my_topic -threads 4 -pause 2 -metrics :2112
```

If required parameters are missing, the app prints usage.

Metrics are available at http://localhost:2112/metrics by default.

Flags (from `src/random_number_generator`):
- -brokers: Comma-separated Kafka brokers (required)
- -topic: Kafka topic (required)
- -threads: Number of producer threads (default: CPU count)
- -pause: Pause per thread between writes in seconds (default: 1)
- -metrics: Address to serve Prometheus metrics (default: :2112)
- -kconfig: Opaque additional Kafka config (logged only)
