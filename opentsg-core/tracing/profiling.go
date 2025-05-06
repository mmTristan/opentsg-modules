package tracing

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	Alloc         = "Alloc"
	TotalAlloc    = "TotalAlloc"
	HeapInUse     = "HeapInUse"
	GCCPUFraction = "GCCPUFraction"
)

/*

write

- pre
- post
- difference
- possibly something else, like a wait group that inremenets eery second.
cumilative usage

*/

// OtelMiddleWare creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// IT adds a profile event that documents the memory usage.
func OtelMiddleWareProfile(ctx context.Context, tracer trace.Tracer) func(h tsg.Handler) tsg.Handler {

	return func(h tsg.Handler) tsg.Handler {
		return tsg.HandlerFunc(func(resp tsg.Response, req *tsg.Request) {

			var memBefore runtime.MemStats
			runtime.ReadMemStats(&memBefore)

			// add some extra spas in
			traceCtx, span := tracer.Start(ctx, req.PatchProperties.WidgetFullID,
				trace.WithAttributes(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()
			// update the context with the trace
			req.Context = traceCtx

			h.Handle(resp, req)
			//attribute.Int(int(memAfter.Alloc - memBefore.Alloc))
			// Capture memory statistics after execution
			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)

			// Choose which stats to add here
			// https://pkg.go.dev/runtime#MemStats
			span.AddEvent("Profile", trace.WithAttributes(

				attribute.KeyValue{Key: Alloc, Value: attribute.IntValue(int(memAfter.Alloc - memBefore.Alloc))},
				attribute.KeyValue{Key: TotalAlloc, Value: attribute.IntValue(int(memAfter.TotalAlloc))},
				attribute.KeyValue{Key: HeapInUse, Value: attribute.IntValue(int(memAfter.HeapInuse))},
				attribute.KeyValue{Key: GCCPUFraction, Value: attribute.Float64Value(memAfter.GCCPUFraction)},
			))

		})
	}
}

// OtelMiddleWare creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// It adds a profile event that documents the memory usage.
func OtelMiddleWarePreProfile(ctx context.Context, tracer trace.Tracer) func(h tsg.Handler) tsg.Handler {

	return func(h tsg.Handler) tsg.Handler {
		return tsg.HandlerFunc(func(resp tsg.Response, req *tsg.Request) {

			var memBefore runtime.MemStats
			runtime.ReadMemStats(&memBefore)

			// add some extra spas in
			traceCtx, span := tracer.Start(ctx, req.PatchProperties.WidgetFullID,
				trace.WithAttributes(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()
			// update the context with the trace
			req.Context = traceCtx

			h.Handle(resp, req)
			//attribute.Int(int(memAfter.Alloc - memBefore.Alloc))
			// Capture memory statistics after execution

			// Choose which stats to add here
			// https://pkg.go.dev/runtime#MemStats
			span.AddEvent("Profile", trace.WithAttributes(

				attribute.KeyValue{Key: Alloc, Value: attribute.IntValue(int(memBefore.Alloc))},
				attribute.KeyValue{Key: TotalAlloc, Value: attribute.IntValue(int(memBefore.TotalAlloc))},
				attribute.KeyValue{Key: HeapInUse, Value: attribute.IntValue(int(memBefore.HeapInuse))},
				attribute.KeyValue{Key: GCCPUFraction, Value: attribute.Float64Value(memBefore.GCCPUFraction)},
			))

		})
	}
}

// OtelMiddleWare creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// It adds a profile event that documents the memory usage.
func OtelMiddleWarePostProfile(ctx context.Context, tracer trace.Tracer) func(h tsg.Handler) tsg.Handler {

	return func(h tsg.Handler) tsg.Handler {
		return tsg.HandlerFunc(func(resp tsg.Response, req *tsg.Request) {

			// add some extra spas in
			traceCtx, span := tracer.Start(ctx, req.PatchProperties.WidgetFullID,
				trace.WithAttributes(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()
			// update the context with the trace
			req.Context = traceCtx

			h.Handle(resp, req)
			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)

			//attribute.Int(int(memAfter.Alloc - memBefore.Alloc))
			// Capture memory statistics after execution

			// Choose which stats to add here
			// https://pkg.go.dev/runtime#MemStats
			span.AddEvent("Profile", trace.WithAttributes(

				attribute.KeyValue{Key: Alloc, Value: attribute.IntValue(int(memAfter.Alloc))},
				attribute.KeyValue{Key: TotalAlloc, Value: attribute.IntValue(int(memAfter.TotalAlloc))},
				attribute.KeyValue{Key: HeapInUse, Value: attribute.IntValue(int(memAfter.HeapInuse))},
				attribute.KeyValue{Key: GCCPUFraction, Value: attribute.Float64Value(memAfter.GCCPUFraction)},
			))

		})
	}
}

// OtelMiddleWare creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// It adds a profile event that documents the memory usage, from
// a calculated average of the memory profile while the handler is running.
// If the duration is small then the profiling will slow down your program.
func OtelMiddleWareAvgProfile(ctx context.Context, tracer trace.Tracer, sampleStep time.Duration) func(h tsg.Handler) tsg.Handler {

	return func(h tsg.Handler) tsg.Handler {
		return tsg.HandlerFunc(func(resp tsg.Response, req *tsg.Request) {

			// add some extra spas in
			traceCtx, span := tracer.Start(ctx, req.PatchProperties.WidgetFullID,
				trace.WithAttributes(),
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()
			// update the context with the trace
			req.Context = traceCtx

			// run the handler as  go function
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				h.Handle(resp, req)
				wg.Done()
			}()

			// collect some stats while it is running
			// @TODO let the user choose the averaging
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			alloc := mem.Alloc
			totalAlloc := mem.TotalAlloc
			heapInUse := mem.HeapInuse
			gCCPUFraction := 0.0
			finish := make(chan bool, 1)
			count := uint64(1)
			go func() {
				monitor := true
				for monitor {

					select {
					case <-finish:
						monitor = false
					case <-time.Tick(sampleStep):
						// sample the memory now
						var mem runtime.MemStats
						runtime.ReadMemStats(&mem)

						alloc = (alloc*count + mem.Alloc) / (count + 1)
						heapInUse = (heapInUse*count + mem.HeapInuse) / (count + 1)
						totalAlloc = mem.TotalAlloc
						gCCPUFraction = (gCCPUFraction*float64(count) + mem.GCCPUFraction) / (float64(count) + 1)

						count++
					}
				}
			}()

			wg.Wait()
			// finish immediately
			finish <- true

			//attribute.Int(int(memAfter.Alloc - memBefore.Alloc))
			// Capture memory statistics after execution

			// Choose which stats to add here
			// https://pkg.go.dev/runtime#MemStats
			span.AddEvent("Profile", trace.WithAttributes(

				attribute.KeyValue{Key: Alloc, Value: attribute.IntValue(int(alloc))},
				attribute.KeyValue{Key: TotalAlloc, Value: attribute.IntValue(int(totalAlloc))},
				attribute.KeyValue{Key: HeapInUse, Value: attribute.IntValue(int(heapInUse))},
				attribute.KeyValue{Key: GCCPUFraction, Value: attribute.Float64Value(gCCPUFraction)},
			))

		})
	}
}
