package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
)

// Random number generation using crypto/rand to avoid math/rand seeding and concurrency issues
func randomInt(min, max int) int {
	if max <= min {
		return min
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback deterministic but safe
		return min + int(time.Now().UnixNano()%int64(max-min+1))
	}
	n := binary.LittleEndian.Uint64(b[:])
	return min + int(n%uint64(max-min+1))
}

type config struct {
	brokers     string
	topic       string
	threads     int
	pauseSec    int
	metricsAddr string
	kconfig     string
	demo        bool
}

var (
	totalWritten   uint64
	numbersWritten = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "numbers_written_total",
		Help: "Total numbers written to Kafka",
	})
	threadStarts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "producer_thread_starts_total",
		Help: "Count of producer thread starts",
	})
	threadEnds = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "producer_thread_ends_total",
		Help: "Count of producer thread ends",
	})
	writeErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "producer_write_errors_total",
		Help: "Count of Kafka write errors",
	})
	writeLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "producer_write_latency_seconds",
		Help:    "Latency for producing a single message",
		Buckets: prometheus.DefBuckets,
	})
	startTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "service_start_time_seconds",
		Help: "Unix time when service started",
	})
	shutdownTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "service_shutdown_time_seconds",
		Help: "Unix time when service shut down",
	})
)

func init() {
	prometheus.MustRegister(numbersWritten, threadStarts, threadEnds, writeErrors, writeLatency, startTime, shutdownTime)
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.brokers, "brokers", "", "Kafka broker list (comma-separated), e.g. localhost:9092")
	flag.StringVar(&cfg.topic, "topic", "", "Kafka topic to produce to")
	flag.IntVar(&cfg.threads, "threads", runtime.NumCPU(), "Number of producer threads")
	flag.IntVar(&cfg.pauseSec, "pause", 1, "Pause between writes per thread (seconds)")
	flag.StringVar(&cfg.metricsAddr, "metrics", ":2112", "Prometheus metrics listen address, e.g. :2112")
	flag.StringVar(&cfg.kconfig, "kconfig", "", "Additional Kafka configuration (opaque string, e.g. key1=val1,key2=val2)")
	flag.BoolVar(&cfg.demo, "demo", false, "Run in demo mode (print random numbers; ignore Kafka flags)")
	flag.Parse()
	return cfg
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [-demo] -brokers <host:port,host:port> -topic <topic> [-threads N] [-pause SECONDS] [-metrics :PORT]\n", os.Args[0])
	fmt.Printf("\nby William Mortl\n\n")
	flag.PrintDefaults()
	fmt.Println("")
}

func validate(cfg config) error {
	if cfg.demo {
		// In demo mode, brokers/topic are not required
		if cfg.threads <= 0 {
			return fmt.Errorf("threads must be > 0")
		}
		if cfg.pauseSec < 0 {
			return fmt.Errorf("pause must be >= 0")
		}
		return nil
	}
	if cfg.brokers == "" || cfg.topic == "" {
		return fmt.Errorf("missing required parameters")
	}
	if cfg.threads <= 0 {
		return fmt.Errorf("threads must be > 0")
	}
	if cfg.pauseSec < 0 {
		return fmt.Errorf("pause must be >= 0")
	}
	return nil
}

func newWriter(brokersCSV, topic string) *kafka.Writer {
	brokers := []string{}
	// Minimal comma split
	for _, b := range splitComma(brokersCSV) {
		if b != "" {
			brokers = append(brokers, b)
		}
	}
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: time.Millisecond * 10,
		Async:        false,
	}
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func producerThread(ctx context.Context, id int, w *kafka.Writer, pause time.Duration, wg *sync.WaitGroup, demo bool) {
	defer wg.Done()
	threadStarts.Inc()
	log.Printf("thread %d: start", id)
	defer func() {
		threadEnds.Inc()
		log.Printf("thread %d: end", id)
	}()

	ticker := time.NewTicker(pause)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// produce/print one message then sleep
		val := randomInt(1, 10000)
		if demo {
			fmt.Printf("%d : %d\n", id, val)
		} else {
			msg := kafka.Message{Value: []byte(fmt.Sprintf("%d", val))}
			start := time.Now()
			if err := w.WriteMessages(ctx, msg); err != nil {
				writeErrors.Inc()
				log.Printf("thread %d: write error: %v", id, err)
			} else {
				numbersWritten.Inc()
				atomic.AddUint64(&totalWritten, 1)
			}
			writeLatency.Observe(time.Since(start).Seconds())
		}

		// sleep
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func main() {
	cfg := parseFlags()
	if err := validate(cfg); err != nil {
		usage()
		os.Exit(2)
	}

	log.Printf("service starting demo=%v brokers=%s topic=%s threads=%d pause=%ds metrics=%s kconfig=%s", cfg.demo, cfg.brokers, cfg.topic, cfg.threads, cfg.pauseSec, cfg.metricsAddr, cfg.kconfig)
	startTime.Set(float64(time.Now().Unix()))

	// Start metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: cfg.metricsAddr, Handler: mux}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics server error: %v", err)
		}
	}()

	// Kafka writer shared across threads (unless demo)
	var writer *kafka.Writer
	if !cfg.demo {
		writer = newWriter(cfg.brokers, cfg.topic)
		defer writer.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(cfg.threads)
	pause := time.Duration(cfg.pauseSec) * time.Second
	for i := 0; i < cfg.threads; i++ {
		go producerThread(ctx, i+1, writer, pause, &wg, cfg.demo)
	}

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Printf("service shutting down")
	shutdownTime.Set(float64(time.Now().Unix()))
	cancel()

	// give threads time to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Printf("timeout waiting for threads to finish")
	}

	// Shutdown metrics server gracefully
	ctxTimeout, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	_ = srv.Shutdown(ctxTimeout)

	log.Printf("service stopped. total numbers written: %d", atomic.LoadUint64(&totalWritten))
}
