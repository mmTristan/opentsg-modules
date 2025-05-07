# Tracing

Tracing uses [OpenTelemetry][OTEL] to run the tracing
of OpenTSG.

## Using Tracing

The tracing provides a way to trace openTSG as it is running,
with low impact middleware plugins that can be used in tandem
with other middlewares.

With the potential for intergration with other tracing services such as jaeger.

This library uses the OpenTelemetry SDK to write to an user defined
io.Writer etc.

### Initialising a Tracer

To start the tracing, you first need to make
a tracer object and to start it.

The tracer is intialised with the following code.

```go
// handle your own error
tracer, closer, _ := InitProvider(nil)
ctx := context.Background()

// run a tracer
// and generate the context with
// the tracer body
c, _ := tracer.Start(ctx, "OpenTSG")

// close the trace at the end of the function
defer close(c)

//Create middlewares from here and run the program
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
/*
Set up tsg and tracer beforehand
*/

// ctx is the context returned when the tracer
// is started
pseudoTSG.Use(OtelMiddleWare(ctx, tracer))
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
/*
Set up tsg and tracer beforehand
*/

// ctx is the context returned when the tracer
// is started
pseudoTSG.Use(tracing.OtelMiddleWareAvgProfile(c, tracer, 100*time.Millisecond))
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

/*
Set up tsg beforehand
*/

pseudoTSG.UseSearches(OtelSearchMiddleWare(tracer))

```

#### Writing to Slog

To write to the `logging/slog`library

```go
tracing.SlogInfoWriter{}
```

can be used as an io.Writer, this writes to the default slog.
At a log level of `slog.LevelInfo`.

[OTEL]: "https://opentelemetry.io/" "The OpenTelemetry website"
