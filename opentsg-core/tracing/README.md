# Tracing

Tracing uses [OpenTelemetry][OTEL] to run the tracing
of OpenTSG.

This library provides the middleware for checking the events that
occur in OpenTSG and their parents that caused them. As well
as offering memory profiling of the widgets.
Allowing you to build up an image of how OpenTSG has run.

## Using Tracing

The tracing provides a way to trace openTSG as it is running,
with low impact middleware plugins that can be used in tandem
with other middlewares.

Get the library with

```cmd
go get "github.com/mrmxf/opentsg-modules/opentsg-core/tracing"
```

OTEL has the potential for integration with other
tracing services such as jaeger, so you can generate more sophisticated
reports from OpenTSG.

This library uses the OpenTelemetry SDK to write to a user
defined location, with the io.Writer interface.



### Initialising a Tracer

To start the tracing, you first need to make
a tracer object and to start it.

The tracer is intialised
and started with the following code.

```go
// handle your own error
tracer, closer, _ := tracing.InitProvider(nil)
ctx := context.Background()

// run a tracer
// and generate the context with
// the tracer body
c, _ := tracer.Start(ctx, "OpenTSG")

// close the trace at the end of the function
defer close(c)

// Create middlewares from here and run the program
// Or manually create a trace
```

The context can be used to intialise any middleware
you would like to use.

This library utilises the openTelemetry SDK of
`"go.opentelemetry.io/otel/sdk/trace"`, which can
be used to customise the tracer object with the
`sdktrace.TracerProviderOption` type.

#### Middleware

Tracing middleware for OpenTSG is provided and can be utilised
with the following code.

```go

import (
    "time"
    "github.com/mrmxf/opentsg-modules/opentsg-core/tracing"
    "github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
)

func main() {
    tracer, closer, _ := tracing.InitProvider(nil)
    ctx := context.Background()

    // run a tracer
    // and generate the context with
    // the tracer body
    c, _ := tracer.Start(ctx, "OpenTSG")

    // close the trace at the end of the function
    defer closer(c)

    otsg, _ := tsg.BuildOpenTSG(commandInputs, *profile, *debug, 
        &tsg.RunnerConfiguration{RunnerCount: 1, ProfilerEnabled: true}, myFlags...)

    // c is the context returned when the tracer
    // is started
    otsg.Use(tracing.OtelMiddleWare(c, tracer))

    // run the engine
    otsg.Run("")

}
```

This logs:

- Start time
- End time
- The job ID
- The openTSG version
- The widget being run

If you want to profile the engine then you can use
one of the profiling middlewares, such as the example
below.

```go
import (
    "time"
    "github.com/mrmxf/opentsg-modules/opentsg-core/tracing"
    "github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
)

func main() {
    tracer, closer, _ := tracing.InitProvider(nil)
    ctx := context.Background()

    // run a tracer
    // and generate the context with
    // the tracer body
    c, _ := tracer.Start(ctx, "OpenTSG")

    // close the trace at the end of the function
    defer closer(c)

    otsg, _ := tsg.BuildOpenTSG(commandInputs, *profile, *debug, 
        &tsg.RunnerConfiguration{RunnerCount: 1, ProfilerEnabled: true}, myFlags...)

    // c is the context returned when the tracer
    // is started
    pseudoTSG.Use(tracing.OtelMiddleWareAvgProfile(c, tracer, 100*time.Millisecond))
    // run the engine
    otsg.Run("")
}
```

This logs:

- Start time
- End time
- The job ID
- The openTSG version
- The widget being run
- The current memory allocation being used in bytes - `Alloc`
- The total memory in bytes used in the lifetime of the program - `TotalAlloc`
- The memory heaps in bytes, in use by the program - `HeapInUse`
- The total percentage of the CPU in use by the program, that is used by the
Garbage Cleaner - `GCCPUFraction`
- The total Bytes used by the heap - `HeapAlloc`
- The number of objects in the heap - `HeapObjects`

#### SearchMiddleware

```go

import (
    "time"
    "github.com/mrmxf/opentsg-modules/opentsg-core/tracing"
    "github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
)

func main() {
    tracer, closer, _ := tracing.InitProvider(nil)
    ctx := context.Background()

    // run a tracer
    // and generate the context with
    // the tracer body
    c, _ := tracer.Start(ctx, "OpenTSG")

    // close the trace at the end of the function
    defer closer(c)

    otsg, _ := tsg.BuildOpenTSG(commandInputs, *profile, *debug, 
        &tsg.RunnerConfiguration{RunnerCount: 1, ProfilerEnabled: true}, myFlags...)

    // Add the tracing middleware
    pseudoTSG.UseSearches(tracing.OtelSearchMiddleWare(tracer))

    // run the engine
    otsg.Run("")
}

```

#### Writing to Slog

To write to the `logging/slog`library

```go
tracing.SlogInfoWriter{}
```

can be used as an io.Writer, this writes to the default slog.
At a log level of `slog.LevelInfo`.

#### Manually creating a Trace

Traces can be created manually as well as from using the
middleware functions provided.

To manually create a trace event
the following example

```go

import (
    "context"
    "go.opentelemetry.io/otel/trace"
)

func ExampleFunc(ctx context.Context, tracer trace.Tracer){
    traceCtx, span := tracer.Start(ctx, "myExampleName",
        trace.WithAttributes(),
        trace.WithSpanKind(trace.SpanKindInternal),
    )
    // end the span at the end of the function
    defer span.End()

    // Write the rest of the function here
}
```

If the context contains previous tracer information,
then the trace will inherit this and make it the parent of that trace.

[OTEL]: "https://opentelemetry.io/" "The OpenTelemetry website"
