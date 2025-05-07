// package tracing contains the OpenTelemetry tracing components
// and middleware
package tracing

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
)

// Configuration sets the tracer output
// and the name of the tracer
type Configuration struct {
	Writer              io.Writer
	InstrumentationName string
}

// Init returns an instance of Jaeger Tracer.
// @TODO finish implementing this
// Use it at your own risk
func InitJaeger(ctx context.Context, service string) trace.Tracer {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
	)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatal("creating OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	//	sdktrace.WithResource(newResource(service)),
	)

	return tp.Tracer(service)
}

// InitProvider sets up the configuration for a OpenTelemtry Tracer
// If conf is nil, then the default writer is to os.Stdout
// and the tracer is given no instrument name.
/*

You can start the tracer with the following code.

	// handle your own error
	tracer, closer, _ := InitProvider(nil)
	ctx := context.Background()

	// run a tracer
	// and generate the context with
	// the tracer body
	c, _ := tracer.Start(ctx, "OpenTSG")

	// close the trace at the end of the function
	defer close(c)

*/
func InitProvider(conf *Configuration, opts ...sdktrace.TracerProviderOption) (trace.Tracer, func(context.Context) error, error) {

	if conf == nil {
		conf = &Configuration{Writer: os.Stdout}
	}
	// default is single line jsons to os.Stdout
	// For choosing your own writers
	// stdouttrace.WithWriter(f)
	// For pretty print
	// stdouttrace.WithPrettyPrint())
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(conf.Writer),
	) //stdouttrace.WithPrettyPrint())

	if err != nil {
		return nil, nil, fmt.Errorf("error creating trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		opts...,
	)

	// register the span processor to stream in realtime
	tracerProvider.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter))
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Tracer(conf.InstrumentationName), tracerProvider.Shutdown, nil

}

// Resource Options contains the fields
// for the resource that is running the tracing.
type ResourceOptions struct {
	ServiceVersion string
	ServiceName    string
	JobID          string
}

// Resources generates the attributes for the tracing,
// giving additional information about the resource doing the tracing.
func Resources(opts *ResourceOptions) sdktrace.TracerProviderOption {

	if opts == nil {
		opts = &ResourceOptions{}
	}

	return sdktrace.WithResource(resourceOpts(*opts))
}

func resourceOpts(opts ResourceOptions) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(opts.ServiceName),
		semconv.ServiceVersion(opts.ServiceVersion),
		attribute.KeyValue{Key: "JobID", Value: attribute.StringValue(opts.JobID)},
	)
}

// SlogInfoWriter enables the io.Writer interface,
// that writes to the default slog object, at
// slog.LevelInfo.
// It can be used to plug slogging into the tracing middleware,
// via the io.Writer interface.
type SlogInfoWriter struct {
}

// Write writes the byte stream as a string to slog.Log
// at slog.LevelInfo. Writing to the default slog writers
func (s SlogInfoWriter) Write(message []byte) (int, error) {
	slog.Log(nil, slog.LevelInfo, string(message))

	return len(message), nil
}

// wrap the writer to extract the messages as well with the logging

// OtelMiddleWare creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
func OtelMiddleWare(ctx context.Context, tracer trace.Tracer) func(h tsg.Handler) tsg.Handler {

	return func(h tsg.Handler) tsg.Handler {
		return tsg.HandlerFunc(func(resp tsg.Response, req *tsg.Request) {

			// add some extra spas in
			traceCtx, span := tracer.Start(ctx, req.PatchProperties.WidgetFullID,
				trace.WithAttributes(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()
			req.Context = traceCtx

			h.Handle(resp, req)
			// span.SetAttributes()
			// @TODO add events for extra information, such as requests etc
			span.AddEvent("test", trace.WithAttributes(
				attribute.KeyValue{Key: "result", Value: attribute.StringValue("tester")},
			))
		})
	}
}

// OtelSearchMiddleware adds middleware to the request search function
func OtelSearchMiddleWare(tracer trace.Tracer) func(tsg.Search) tsg.Search {

	return func(search tsg.Search) tsg.Search {

		return tsg.SearchFunc(func(ctx context.Context, URI string) ([]byte, error) {

			_, span := tracer.Start(ctx, URI,

				trace.WithAttributes(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()

			return search.Search(ctx, URI)
		})
	}
}
