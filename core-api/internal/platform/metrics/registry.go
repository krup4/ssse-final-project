package metrics

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

var latencyBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type Registry struct {
	mu sync.RWMutex

	httpInFlight int64
	httpRequests map[httpKey]int64
	httpErrors   map[httpKey]int64
	httpLatency  map[httpKey]*histogram

	kafkaMessages  map[kafkaKey]int64
	kafkaErrors    map[kafkaKey]int64
	kafkaCommits   map[kafkaKey]int64
	kafkaPublishes map[kafkaKey]int64
}

type httpKey struct {
	Method string
	Route  string
	Status string
}

type kafkaKey struct {
	Topic  string
	Result string
}

type histogram struct {
	Buckets []int64
	Sum     float64
	Count   int64
}

func NewRegistry() *Registry {
	return &Registry{
		httpRequests:   map[httpKey]int64{},
		httpErrors:     map[httpKey]int64{},
		httpLatency:    map[httpKey]*histogram{},
		kafkaMessages:  map[kafkaKey]int64{},
		kafkaErrors:    map[kafkaKey]int64{},
		kafkaCommits:   map[kafkaKey]int64{},
		kafkaPublishes: map[kafkaKey]int64{},
	}
}

func (r *Registry) IncHTTPInFlight() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.httpInFlight++
}

func (r *Registry) DecHTTPInFlight() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.httpInFlight > 0 {
		r.httpInFlight--
	}
}

func (r *Registry) ObserveHTTPRequest(method, route string, status int, duration time.Duration) {
	if route == "" {
		route = "unmatched"
	}
	key := httpKey{Method: method, Route: route, Status: fmt.Sprintf("%d", status)}
	seconds := duration.Seconds()

	r.mu.Lock()
	defer r.mu.Unlock()
	r.httpRequests[key]++
	if status >= 500 {
		r.httpErrors[key]++
	}
	hist := r.httpLatency[key]
	if hist == nil {
		hist = &histogram{Buckets: make([]int64, len(latencyBuckets))}
		r.httpLatency[key] = hist
	}
	for i, bucket := range latencyBuckets {
		if seconds <= bucket {
			hist.Buckets[i]++
		}
	}
	hist.Sum += seconds
	hist.Count++
}

func (r *Registry) IncKafkaMessage(topic, result string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kafkaMessages[kafkaKey{Topic: topic, Result: result}]++
}

func (r *Registry) IncKafkaError(topic, result string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kafkaErrors[kafkaKey{Topic: topic, Result: result}]++
}

func (r *Registry) IncKafkaCommit(topic, result string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kafkaCommits[kafkaKey{Topic: topic, Result: result}]++
}

func (r *Registry) IncKafkaPublish(topic, result string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kafkaPublishes[kafkaKey{Topic: topic, Result: result}]++
}

func (r *Registry) WritePrometheus(w io.Writer) {
	snapshot := r.snapshot()
	writeGauge(w, "core_api_http_requests_in_flight", "Current HTTP requests in flight.", float64(snapshot.httpInFlight), nil)
	writeCounterMap(w, "core_api_http_requests_total", "Total HTTP requests.", snapshot.httpRequests)
	writeCounterMap(w, "core_api_http_errors_total", "Total HTTP 5xx responses.", snapshot.httpErrors)
	writeHTTPHistogram(w, "core_api_http_request_duration_seconds", "HTTP request latency in seconds.", snapshot.httpLatency)
	writeKafkaCounterMap(w, "core_api_kafka_messages_total", "Total Kafka messages processed by result.", snapshot.kafkaMessages)
	writeKafkaCounterMap(w, "core_api_kafka_errors_total", "Total Kafka consumer errors by result.", snapshot.kafkaErrors)
	writeKafkaCounterMap(w, "core_api_kafka_commits_total", "Total Kafka offset commits by result.", snapshot.kafkaCommits)
	writeKafkaCounterMap(w, "core_api_kafka_publishes_total", "Total Kafka producer publishes by result.", snapshot.kafkaPublishes)
}

type snapshot struct {
	httpInFlight   int64
	httpRequests   map[httpKey]int64
	httpErrors     map[httpKey]int64
	httpLatency    map[httpKey]histogram
	kafkaMessages  map[kafkaKey]int64
	kafkaErrors    map[kafkaKey]int64
	kafkaCommits   map[kafkaKey]int64
	kafkaPublishes map[kafkaKey]int64
}

func (r *Registry) snapshot() snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	httpLatency := make(map[httpKey]histogram, len(r.httpLatency))
	for key, hist := range r.httpLatency {
		httpLatency[key] = histogram{
			Buckets: append([]int64(nil), hist.Buckets...),
			Sum:     hist.Sum,
			Count:   hist.Count,
		}
	}
	return snapshot{
		httpInFlight:   r.httpInFlight,
		httpRequests:   copyHTTPMap(r.httpRequests),
		httpErrors:     copyHTTPMap(r.httpErrors),
		httpLatency:    httpLatency,
		kafkaMessages:  copyKafkaMap(r.kafkaMessages),
		kafkaErrors:    copyKafkaMap(r.kafkaErrors),
		kafkaCommits:   copyKafkaMap(r.kafkaCommits),
		kafkaPublishes: copyKafkaMap(r.kafkaPublishes),
	}
}

func copyHTTPMap(in map[httpKey]int64) map[httpKey]int64 {
	out := make(map[httpKey]int64, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func copyKafkaMap(in map[kafkaKey]int64) map[kafkaKey]int64 {
	out := make(map[kafkaKey]int64, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func writeGauge(w io.Writer, name, help string, value float64, labels map[string]string) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s gauge\n%s%s %g\n", name, help, name, name, formatLabels(labels), value)
}

func writeCounterMap(w io.Writer, name, help string, values map[httpKey]int64) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s counter\n", name, help, name)
	keys := make([]httpKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return httpKeyLess(keys[i], keys[j]) })
	for _, key := range keys {
		labels := map[string]string{"method": key.Method, "route": key.Route, "status": key.Status}
		fmt.Fprintf(w, "%s%s %d\n", name, formatLabels(labels), values[key])
	}
}

func writeKafkaCounterMap(w io.Writer, name, help string, values map[kafkaKey]int64) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s counter\n", name, help, name)
	keys := make([]kafkaKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Topic != keys[j].Topic {
			return keys[i].Topic < keys[j].Topic
		}
		return keys[i].Result < keys[j].Result
	})
	for _, key := range keys {
		labels := map[string]string{"topic": key.Topic, "result": key.Result}
		fmt.Fprintf(w, "%s%s %d\n", name, formatLabels(labels), values[key])
	}
}

func writeHTTPHistogram(w io.Writer, name, help string, values map[httpKey]histogram) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s histogram\n", name, help, name)
	keys := make([]httpKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return httpKeyLess(keys[i], keys[j]) })
	for _, key := range keys {
		hist := values[key]
		for i, bucket := range latencyBuckets {
			labels := map[string]string{"method": key.Method, "route": key.Route, "status": key.Status, "le": fmt.Sprintf("%g", bucket)}
			fmt.Fprintf(w, "%s_bucket%s %d\n", name, formatLabels(labels), hist.Buckets[i])
		}
		labels := map[string]string{"method": key.Method, "route": key.Route, "status": key.Status, "le": "+Inf"}
		fmt.Fprintf(w, "%s_bucket%s %d\n", name, formatLabels(labels), hist.Count)
		baseLabels := map[string]string{"method": key.Method, "route": key.Route, "status": key.Status}
		fmt.Fprintf(w, "%s_sum%s %g\n", name, formatLabels(baseLabels), hist.Sum)
		fmt.Fprintf(w, "%s_count%s %d\n", name, formatLabels(baseLabels), hist.Count)
	}
}

func httpKeyLess(a, b httpKey) bool {
	if a.Route != b.Route {
		return a.Route < b.Route
	}
	if a.Method != b.Method {
		return a.Method < b.Method
	}
	return a.Status < b.Status
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, escapeLabel(labels[key])))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
