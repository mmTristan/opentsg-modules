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
	HeapAlloc     = "HeapAlloc"
	HeapObjects   = "HeapObjects"
)

// OtelMiddleWarePreProfile creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// It adds a profile event that documents the memory usage, before the handler is run.
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

				attribute.Int(Alloc, int(memBefore.Alloc)),
				attribute.Int(TotalAlloc, int(memBefore.TotalAlloc)),
				attribute.Int(HeapInUse, int(memBefore.HeapInuse)),
				attribute.Float64(GCCPUFraction, memBefore.GCCPUFraction),
				attribute.Int(HeapObjects, int(memBefore.HeapObjects)),
				attribute.Int(HeapAlloc, int(memBefore.HeapAlloc)),
			))

		})
	}
}

// OtelMiddleWarePostProfile creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// It adds a profile event that documents the memory usage, after the handler is run.
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

				attribute.Int(Alloc, int(memAfter.Alloc)),
				attribute.Int(TotalAlloc, int(memAfter.TotalAlloc)),
				attribute.Int(HeapInUse, int(memAfter.HeapInuse)),
				attribute.Float64(GCCPUFraction, memAfter.GCCPUFraction),
				attribute.Int(HeapObjects, int(memAfter.HeapObjects)),
				attribute.Int(HeapAlloc, int(memAfter.HeapAlloc)),
			))

		})
	}
}

// OtelMiddleWareAvgPRofile creates an openTelemetry middleware that uses the tracer,
// Documenting the run time of the widget and the version of OTSG used.
// It adds a profile event that documents the memory usage, from
// a calculated average of the memory profile while the handler is running.
// If the duration is a small increment then the profiling will slow down your program.
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
			heapAlloc := mem.HeapAlloc
			heapObjects := mem.HeapObjects
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

						totalCount := count + 1

						alloc = (alloc*count + mem.Alloc) / (totalCount)
						heapInUse = (heapInUse*count + mem.HeapInuse) / (totalCount)
						totalAlloc = mem.TotalAlloc
						gCCPUFraction = (gCCPUFraction*float64(count) + mem.GCCPUFraction) / (float64(totalCount))
						heapAlloc = (heapAlloc*count + mem.HeapAlloc) / (totalCount)
						heapObjects = (heapObjects*count + mem.HeapObjects) / (totalCount)

						count = totalCount
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

				attribute.Int(Alloc, int(alloc)),
				attribute.Int(TotalAlloc, int(totalAlloc)),
				attribute.Int(HeapInUse, int(heapInUse)),
				attribute.Float64(GCCPUFraction, gCCPUFraction),
				attribute.Int(HeapAlloc, int(heapAlloc)),
				attribute.Int(HeapObjects, int(heapObjects)),
			))

		})
	}
}
