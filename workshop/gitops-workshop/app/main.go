package main

import (
	"context"
	"encoding/json"
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
)

type metricsStore struct {
	startTime       time.Time
	inflight        atomic.Int64
	totalRequests   atomic.Uint64
	totalLatencyMs  atomic.Uint64
	pathCounters    map[string]*atomic.Uint64
	pathCountersMux sync.RWMutex
}

func newMetricsStore(paths []string) *metricsStore {
	m := &metricsStore{
		startTime:    time.Now(),
		pathCounters: make(map[string]*atomic.Uint64),
	}
	for _, p := range paths {
		m.pathCounters[p] = &atomic.Uint64{}
	}
	m.pathCounters["other"] = &atomic.Uint64{}
	return m
}

func (m *metricsStore) incPath(path string) {
	m.pathCountersMux.RLock()
	counter, ok := m.pathCounters[path]
	m.pathCountersMux.RUnlock()
	if !ok {
		counter = m.pathCounters["other"]
	}
	counter.Add(1)
}

func (m *metricsStore) record(latencyMs uint64, path string) {
	m.totalRequests.Add(1)
	m.totalLatencyMs.Add(latencyMs)
	m.incPath(path)
}

func (m *metricsStore) render(w http.ResponseWriter) {
	uptime := time.Since(m.startTime).Seconds()
	totalReq := m.totalRequests.Load()
	avgLatency := float64(0)
	if totalReq > 0 {
		avgLatency = float64(m.totalLatencyMs.Load()) / float64(totalReq)
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP app_requests_total Total HTTP requests received\n")
	fmt.Fprintf(w, "# TYPE app_requests_total counter\n")
	for path, counter := range m.pathCounters {
		fmt.Fprintf(w, "app_requests_total{path=\"%s\"} %d\n", path, counter.Load())
	}

	fmt.Fprintf(w, "# HELP app_request_latency_ms_average Average latency of handled requests in milliseconds\n")
	fmt.Fprintf(w, "# TYPE app_request_latency_ms_average gauge\n")
	fmt.Fprintf(w, "app_request_latency_ms_average %.3f\n", avgLatency)

	fmt.Fprintf(w, "# HELP app_uptime_seconds Seconds since service start\n")
	fmt.Fprintf(w, "# TYPE app_uptime_seconds gauge\n")
	fmt.Fprintf(w, "app_uptime_seconds %.0f\n", uptime)

	fmt.Fprintf(w, "# HELP app_inflight_requests In-flight HTTP requests\n")
	fmt.Fprintf(w, "# TYPE app_inflight_requests gauge\n")
	fmt.Fprintf(w, "app_inflight_requests %d\n", m.inflight.Load())

	fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n")
	fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytes += n
	return n, err
}

func loggingMiddleware(version string, m *metricsStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.inflight.Add(1)
		defer m.inflight.Add(-1)

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		latency := time.Since(start).Milliseconds()
		m.record(uint64(latency), r.URL.Path)

		entry := map[string]interface{}{
			"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     lrw.statusCode,
			"bytes":      lrw.bytes,
			"latency_ms": latency,
			"version":    version,
		}
		msg, _ := json.Marshal(entry)
		log.Println(string(msg))
	})
}

func main() {
	message := getenv("MESSAGE", "Hello DevOps Workshop")
	version := getenv("VERSION", "v1.0.0")
	port := getenv("PORT", "8080")
	addr := ":" + port

	metrics := newMetricsStore([]string{"/", "/healthz", "/readyz", "/metrics"})

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pod := getenv("HOSTNAME", "unknown")
		resp := map[string]string{
			"message": message,
			"version": version,
			"pod":     pod,
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.render(w)
	})

	handler := loggingMiddleware(version, metrics, mux)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutdown signal received, draining...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped cleanly")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(data)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
