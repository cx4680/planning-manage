package trace

// TraceName represents the tracing name.
const TraceName = ""

const (
	EnvServiceName = "SERVICE_NAME"
)

// A Config is a opentelemetry config.
type Config struct {
	Name     string  `json:",optional"`
	Endpoint string  `json:",optional"`
	Sampler  float64 `json:",default=1.0"`
	Batcher  string  `json:",default=jaeger,options=jaeger|zipkin"`
}
