package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"go.opentelemetry.io/otel"

	. "github.com/smartystreets/goconvey/convey"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

/*

create a test handler to generate some logs

log testing needs to be parsed etc into a buffer then parsed
find the object to parse it into to check it ran

What metrics do I want to check


Check the context span IDs,


print the context and see what was stored when it was run

Tests, ensure the middleware works

test the middleware
and test the making one
*/

type testHandler struct {
}

func (t testHandler) Handle(resp tsg.Response, _ *tsg.Request) {
	resp.Write(tsg.SaveSuccess, "succesful run for a test")
}

func TestTracerInit(t *testing.T) {

	// create the tracer
	buf := bytes.NewBuffer([]byte{})

	trc, end, _ := InitProvider(
		&Configuration{Writer: buf, InstrumentationName: "TestWriter"},
		Resources(&ResourceOptions{
			ServiceName: "Demo Tracer",
		}))
	ctx := context.Background()

	// run a trace
	_, spn := trc.Start(ctx, "test")
	spn.End()

	// flush the buffer
	end(ctx)

	var stub tracetest.SpanStub
	json.Unmarshal(buf.Bytes(), &stub)

	// make some useful comparisons instead of printing it
	fmt.Println(stub)

}

func TestMiddleware(t *testing.T) {

	// Create an in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()

	// Set up a tracer provider
	// Set up a tracer provider with the in-memory exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer := otel.Tracer("example")

	// Start a span

	handler := OtelMiddleWare(nil, tracer)(testHandler{})
	handler.Handle(&tsg.TestResponder{}, &tsg.Request{})

	spans := exporter.GetSpans()
	fmt.Println(exporter)
	Convey("Checking that the middleware succesfully generates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run", func() {
				So(len(spans), ShouldResemble, 1)
				So(spans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

	exporter2 := tracetest.NewInMemoryExporter()
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter2)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer = otel.Tracer("example")

	handler2 := OtelMiddleWareProfile(nil, tracer)(testHandler{})
	resp := &tsg.TestResponder{}
	handler2.Handle(resp, &tsg.Request{})
	fmt.Println(resp)

	// set up more profile tests
	spans = exporter2.GetSpans()
	by, _ := json.Marshal(spans[0])
	fmt.Println(string(by))
	fmt.Println(spans[0].Events[0].Name)
	Convey("Checking that the middleware succesfully generates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run", func() {
				So(len(spans), ShouldResemble, 1)
				So(len(spans[0].Events), ShouldResemble, 1)
				So(spans[0].Events[0].Name, ShouldResemble, "Profile")
				So(spans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

	exporter3 := tracetest.NewInMemoryExporter()
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter3)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer = otel.Tracer("example")

	handler3 := OtelMiddleWareAvgProfile(nil, tracer, time.Nanosecond)(testHandler{})
	resp = &tsg.TestResponder{}
	handler3.Handle(resp, &tsg.Request{})
	fmt.Println(resp)

	// set up more profile tests
	spans = exporter3.GetSpans()
	by, _ = json.Marshal(spans[0])
	fmt.Println(string(by))
	fmt.Println(spans[0].Events[0].Name)
	Convey("Checking that the middleware succesfully generates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run", func() {
				So(len(spans), ShouldResemble, 1)
				So(len(spans[0].Events), ShouldResemble, 1)
				So(spans[0].Events[0].Name, ShouldResemble, "Profile")
				So(spans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

}

func TestSearchMiddleware(t *testing.T) {

	// Create an in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()

	// Set up a tracer provider
	// Set up a tracer provider with the in-memory exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer := otel.Tracer("example")

	// Start a span

	handler := OtelSearchMiddleWare(tracer)(tsg.SearchFunc(func(_ context.Context, URI string) ([]byte, error) {

		return nil, nil
	}))

	handler.Search(nil, "test")

	spans := exporter.GetSpans()

	Convey("Checking that the search middleware succesfully generates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run", func() {
				So(len(spans), ShouldResemble, 1)
				So(spans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

}

func TestMiddlewareInteraction(t *testing.T) {

	// Create an in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()

	// Set up a tracer provider
	// Set up a tracer provider with the in-memory exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer := otel.Tracer("example")

	// Start a span

	pseudoTSG, err := tsg.BuildOpenTSG("../tsg/testdata/handlerLoaders/singleloader.json", "", true, nil)

	ctx := context.Background()
	pseudoTSG.Use(OtelMiddleWare(ctx, tracer))
	pseudoTSG.UseSearches(OtelSearchMiddleWare(tracer))

	pseudoTSG.HandleFunc("test.fill", tsg.HandlerFunc(func(_ tsg.Response, r *tsg.Request) {
		fmt.Println(r.Context)
		r.SearchWithCredentials(r.Context, "Valid Middleware search")
	}))

	pseudoTSG.Run("")
	spans := exporter.GetSpans()

	b, _ := json.MarshalIndent(spans, "", "    ")

	fmt.Println(string(b))

	var SearchSpan tracetest.SpanStub
	var BlueSpan tracetest.SpanStub
	for _, span := range spans {
		switch span.Name {
		case "Valid Middleware search":
			SearchSpan = span
		case "blue":
			BlueSpan = span
		}

	}
	Convey("Checking that the middleware updates the context of the request", t, func() {
		Convey("checking the search URI log has a parent of the blue widget that called it", func() {
			Convey("The parent contexts match allowing the call to be traced", func() {
				So(err, ShouldBeNil)
				So(SearchSpan.Parent, ShouldResemble, BlueSpan.SpanContext)
			})
		})
	})

}
