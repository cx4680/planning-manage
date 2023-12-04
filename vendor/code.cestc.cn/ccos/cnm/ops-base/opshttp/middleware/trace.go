package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"code.cestc.cn/ccos/cnm/ops-base/trace"
)

const (
	tracerKey = "otel-go-contrib-tracer"
)

func SetTrace() gin.HandlerFunc {
	trace.Init()
	tracer := trace.GetTracer()
	propagators := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		c.Set(tracerKey, tracer)
		savedCtx := c.Request.Context()
		ctx := propagators.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("", c.Request.URL.Path, c.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}
		spanName := c.FullPath()
		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		var traceId, spanId string
		if sc := span.SpanContext(); sc.HasTraceID() {
			traceId = sc.TraceID().String()
			spanId = sc.SpanID().String()

			// 设置span到ctx
			c.Set(trace.CtxTraceSpanKey, trace.NewSpan(traceId, spanId))
			tracerCtx := trace.GetContext(ctx, traceId, spanId)
			c.Request = c.Request.WithContext(tracerCtx)
		}
		c.Writer.Header().Set(trace.HeaderTraceIdKey, traceId)
		c.Writer.Header().Set(trace.HeaderSpanIdKey, spanId)

		// response
		writer := &responseWriter{
			c.Writer,
			bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer

		// request
		var reqBody string
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch ||
			c.Request.Method == http.MethodDelete {

			requestBody, _ := c.GetRawData()
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
			reqBody = string(requestBody)
		}

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(status)...)
		}
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}

		replyBody := writer.b.String()
		span.SetAttributes(attribute.String("request", reqBody))
		span.SetAttributes(attribute.String("response", replyBody))
	}
}
