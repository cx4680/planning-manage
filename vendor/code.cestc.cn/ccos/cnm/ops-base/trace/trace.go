package trace

import (
	"context"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type ContextKey string

const (
	CtxTraceSpanKey = "CtxTraceSpan"
	CtxTraceSpan    = "CtxTrace"
)

var HeaderTraceIdKey = http.CanonicalHeaderKey("x-trace-id")
var HeaderSpanIdKey = http.CanonicalHeaderKey("x-span-id")

type Tracer struct {
	traceId string
	spanId  string
}

func Init() {
	StartAgent(Config{
		Name: os.Getenv(EnvServiceName),
		// TODO 增加jaeger服务
		//Sampler:  1.0,
		//Batcher:  "jaeger",
		//Endpoint: "http://localhost:14268/api/traces",
	})
}

func GetTracer() oteltrace.Tracer {
	tracerProvider := otel.GetTracerProvider()
	host, _ := os.Hostname()
	tracer := tracerProvider.Tracer(
		host,
	)
	return tracer
}

func NewSpan(traceId, spanId string) Tracer {
	return Tracer{
		traceId: traceId,
		spanId:  spanId,
	}
}

func (t *Tracer) Trace() string {
	return t.traceId
}

func (t *Tracer) Span() string {
	return t.spanId
}

func GetContext(ctx context.Context, traceId, spanId string) context.Context {
	ctx = context.WithValue(ctx, CtxTraceSpanKey, NewSpan(traceId, spanId))
	return ctx
}

func GetSpan(ctx context.Context) Tracer {
	span, ok := ctx.Value(CtxTraceSpanKey).(Tracer)
	if ok {
		return span
	}
	return Tracer{}
}

// 设置trace_id
func GenTrace(ctx context.Context, name string) context.Context {
	tracer := otel.GetTracerProvider().Tracer(name)
	_, span := tracer.Start(
		ctx,
		name,
	)

	defer span.End()
	if sc := span.SpanContext(); sc.HasTraceID() {
		return GetContext(ctx, sc.TraceID().String(), sc.SpanID().String())
	}
	return ctx
}

func ExtraTraceID(ctx context.Context) string {
	var traceId string
	span, ok := ctx.Value(string(CtxTraceSpanKey)).(Tracer)
	if ok {
		traceId = span.Trace()
	}
	return traceId
}
