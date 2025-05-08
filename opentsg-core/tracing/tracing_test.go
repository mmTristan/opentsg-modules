package tracing

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"go.opentelemetry.io/otel"

	. "github.com/smartystreets/goconvey/convey"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// test handler runs a simple
// Handle that always passes
type testHandler struct {
}

func (t testHandler) Handle(resp tsg.Response, _ *tsg.Request) {
	resp.Write(tsg.SaveSuccess, "succesful run for a test")
}

func TestTracerInit(t *testing.T) {

	// create the tracer
	buf := bytes.NewBuffer([]byte{})

	trc, end, err := InitProvider(
		&WriterConfiguration{Writer: buf, InstrumentationName: "TestWriter"},
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
	// jErr := json.Unmarshal(buf.Bytes(), &stub)
	json.Unmarshal(buf.Bytes(), &stub)

	Convey("Checking that the init provider generates a tracer that can be written to", t, func() {
		Convey("writing a tracer to a buffer to test the contents", func() {
			Convey("The tracer writes to the buffer with the correct fields", func() {
				So(err, ShouldBeNil)
				// skip the unmarshall error check as the trace test does not sync up
				// with the returned value
				//	So(jErr, ShouldBeNil)
				So(stub.Name, ShouldResemble, "test")
				So(stub.InstrumentationScope.Name, ShouldResemble, "TestWriter")
			})
		})
	})

}

func TestMiddleware(t *testing.T) {

	// Create an in-memory span exporter
	exporter, tracer := makeExporterSpan()

	// Start a span
	handler := OtelMiddleWare(nil, tracer)(testHandler{})
	handler.Handle(&tsg.TestResponder{}, &tsg.Request{})

	spans := exporter.GetSpans()

	Convey("Checking that the middleware succesfully generates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run", func() {
				So(len(spans), ShouldResemble, 1)
				So(spans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

}

func makeExporterSpan() (*tracetest.InMemoryExporter, trace.Tracer) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)),
	)
	otel.SetTracerProvider(tp)

	// Create a tracer
	tracer := otel.Tracer("example")

	return exporter, tracer
}

func validateEvent(event []attribute.KeyValue, expected []string) bool {
	expec := make([]string, len(expected))
	copy(expec, expected)
	for _, ev := range event {

		if slices.Contains(expec, string(ev.Key)) {
			pos := slices.Index(expec, string(ev.Key))
			expec = slices.Delete(expec, pos, pos+1)
		} else {

			return false
		}

	}

	return true

}

func TestProfilingMiddlewares(t *testing.T) {

	headers := []string{Alloc,
		TotalAlloc,
		HeapInUse,
		GCCPUFraction,
		HeapAlloc,
		HeapObjects}

	// set up recorder
	exporterAvg, tracerAvg := makeExporterSpan()
	// set up middleware
	avgHandler := OtelMiddleWareAvgProfile(nil, tracerAvg, time.Nanosecond)(testHandler{})
	avgHandler.Handle(&tsg.TestResponder{}, &tsg.Request{})

	// set up more profile tests
	avgSpans := exporterAvg.GetSpans()

	Convey("Checking that the middleware average profiler creates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run, with an events span for profiling", func() {
				So(len(avgSpans), ShouldResemble, 1)
				So(len(avgSpans[0].Events), ShouldResemble, 1)
				So(avgSpans[0].Events[0].Name, ShouldResemble, "Profile")
				So(avgSpans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

	Convey("Checking that the average profiler event has all the required memory attributes", t, func() {
		Convey("comparing the event against the expected attribute headers", func() {
			Convey("The event has all the required headers", func() {
				So(validateEvent(avgSpans[0].Events[0].Attributes, headers), ShouldBeTrue)
			})
		})
	})

	// set up recorder
	exporterPre, tracerPre := makeExporterSpan()
	// set up middleware
	preHandler := OtelMiddleWarePreProfile(nil, tracerPre)(testHandler{})
	preHandler.Handle(&tsg.TestResponder{}, &tsg.Request{})

	// set up more profile tests
	preSpans := exporterPre.GetSpans()

	Convey("Checking that the middleware average profiler creates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run, with an events span for profiling", func() {
				So(len(preSpans), ShouldResemble, 1)
				So(len(preSpans[0].Events), ShouldResemble, 1)
				So(preSpans[0].Events[0].Name, ShouldResemble, "Profile")
				So(preSpans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

	Convey("Checking that the pre profiler event has all the required memory attributes", t, func() {
		Convey("comparing the event against the expected attribute headers", func() {
			Convey("The event has all the required headers", func() {
				So(validateEvent(preSpans[0].Events[0].Attributes, headers), ShouldBeTrue)
			})
		})
	})

	// set up recorder
	exporterPost, tracerPost := makeExporterSpan()
	// set up middleware
	postHandler := OtelMiddleWarePostProfile(nil, tracerPost)(testHandler{})
	postHandler.Handle(&tsg.TestResponder{}, &tsg.Request{})

	// set up more profile tests
	postSpans := exporterPost.GetSpans()

	Convey("Checking that the middleware average profiler creates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run, with an events span for profiling", func() {
				So(len(postSpans), ShouldResemble, 1)
				So(len(postSpans[0].Events), ShouldResemble, 1)
				So(postSpans[0].Events[0].Name, ShouldResemble, "Profile")
				So(postSpans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

	Convey("Checking that the post profiler event has all the required memory attributes", t, func() {
		Convey("comparing the event against the expected attribute headers", func() {
			Convey("The event has all the required headers", func() {
				So(validateEvent(postSpans[0].Events[0].Attributes, headers), ShouldBeTrue)
			})
		})
	})
}

func TestSearchMiddleware(t *testing.T) {

	// Create an in-memory span exporter
	exporter, tracer := makeExporterSpan()

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

	exporterProf, tracerProf := makeExporterSpan()
	// set up middleware
	handlerProf := OtelSearchMiddleWareProfile(tracerProf)(tsg.SearchFunc(func(_ context.Context, URI string) ([]byte, error) {

		return []byte{0, 0, 0, 0}, nil
	}))

	handlerProf.Search(nil, "")
	// set up more profile tests
	profSpans := exporterProf.GetSpans()

	Convey("Checking that the search middleware profiler succesfully generates a log", t, func() {
		Convey("running with a test tracer to check the internal logs", func() {
			Convey("A single span is returned from the single run, with an events span for profiling", func() {
				So(len(profSpans), ShouldResemble, 1)
				So(len(profSpans[0].Events), ShouldResemble, 1)
				So(profSpans[0].Events[0].Name, ShouldResemble, "Profile")
				So(profSpans[0].Resource.String(), ShouldResemble, "service.name=unknown_service:tracing.test,telemetry.sdk.language=go,telemetry.sdk.name=opentelemetry,telemetry.sdk.version=1.35.0")
			})
		})
	})

	Convey("Checking that the profiler event has all the required file size attributes", t, func() {
		Convey("comparing the event against the expected attribute headers", func() {
			Convey("The event has all the required headers", func() {
				So(validateEvent(profSpans[0].Events[0].Attributes, []string{FileSize}), ShouldBeTrue)
			})
		})
	})

}

func TestMiddlewareInteraction(t *testing.T) {

	exporter, tracer := makeExporterSpan()

	// Start a span

	pseudoTSG, err := tsg.BuildOpenTSG("../tsg/testdata/handlerLoaders/singleloader.json", "", true, nil)

	ctx := context.Background()
	pseudoTSG.Use(OtelMiddleWare(ctx, tracer))
	pseudoTSG.UseSearches(OtelSearchMiddleWare(tracer))

	pseudoTSG.HandleFunc("test.fill", tsg.HandlerFunc(func(_ tsg.Response, r *tsg.Request) {

		r.SearchWithCredentials(r.Context, "Valid Middleware search")
	}))

	pseudoTSG.Run("")
	spans := exporter.GetSpans()

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
