package jaeger

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/DistributedTraceProject/config"
	"github.com/DistributedTraceProject/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	parentKey = "traceparent"
)

type Jaeger struct{}

func init() {
	trace.RegisterProviders(trace.Jaeger, &Jaeger{})
}

func (j *Jaeger) Initialize(_ context.Context, traceCfg *config.Trace) error {
	tp, err := j.tracerProvider(traceCfg)
	if err != nil {
		return err
	}
	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(&propagation.TraceContext{})
	return nil
}

func (j *Jaeger) tracerProvider(traceCfg *config.Trace) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(traceCfg.Address)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(trace.Service),
		)),
	)
	return tp, nil
}

func (j *Jaeger) Extract(ctx *config.Context, hints []*config.Hint) bool {
	var traceContext string
	for _, h := range hints {
		if h.Type != config.TypeTrace {
			continue
		}
		traceContext = h.Inputs[0].V
		break
	}
	if len(traceContext) == 0 {
		return false
	}
	ctx.Context = otel.GetTextMapPropagator().Extract(ctx.Context, propagation.MapCarrier{parentKey: traceContext})
	return true
}
