# Tracing

Tracing uses [OpenTelemetry][OTEL] to run the tracing
of OpenTSG.

## Using Tracing

### Initing a Tracer

Make a tracer with the following code.

Add custom bits to it with `"go.opentelemetry.io/otel/sdk/trace"`
to update the fields and resource you want to track.
Setting up resources with the `Resources` function,
to describe the engine 

#### Middleware

Set up the Middleware with the following example.

```go
/*
Set up tsg beforehand
*/


pseudoTSG.Use(OtelMiddleWare(ctx, tracer))
```

This logs:
- Start time
- End time
- other custom fields

#### SearchMiddleware

```go

/*
Set up tsg beforehand
*/

pseudoTSG.UseSearches(OtelSearchMiddleWare(tracer))

```

#### Writing to Slog

[OTEL]: "https://opentelemetry.io/" "The OpenTelemetry website"
