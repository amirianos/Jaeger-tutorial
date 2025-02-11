
package main

import (
        "context"
        "fmt"
        "io"
        "log"
        "math/rand"
        "net/http"
        "time"

        "github.com/opentracing/opentracing-go"
        "github.com/opentracing/opentracing-go/ext"
        "github.com/uber/jaeger-client-go"
        jaegercfg "github.com/uber/jaeger-client-go/config"
)

func initTracer(serviceName string) (opentracing.Tracer, io.Closer, error) {
        cfg := jaegercfg.Configuration{
                ServiceName: serviceName,
                Sampler: &jaegercfg.SamplerConfig{
                        Type:  jaeger.SamplerTypeConst,
                        Param: 1,
                },
                Reporter: &jaegercfg.ReporterConfig{
                        LogSpans:          true,
                        CollectorEndpoint: "http://localhost:14268/api/traces",
                },
        }
        return cfg.NewTracer()
}

func simulateWork() {
        time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
}

func callServiceC(ctx context.Context) {
        span, _ := opentracing.StartSpanFromContext(ctx, "callServiceC")
        defer span.Finish()

        req, _ := http.NewRequest("GET", "http://localhost:8082/service-c", nil)
        ext.SpanKindRPCClient.Set(span)
        ext.HTTPUrl.Set(span, req.URL.String())
        ext.HTTPMethod.Set(span, "GET")
        opentracing.GlobalTracer().Inject(
                span.Context(),
                opentracing.HTTPHeaders,
                opentracing.HTTPHeadersCarrier(req.Header),
        )

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                log.Fatalf("Error calling Service C: %v", err)
        }
        defer resp.Body.Close()

        fmt.Println("Service C responded with status:", resp.Status)
}

func handler(w http.ResponseWriter, r *http.Request) {
        spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
        span := opentracing.StartSpan("serviceB", ext.RPCServerOption(spanCtx))
        defer span.Finish()

        ctx := opentracing.ContextWithSpan(context.Background(), span)

        simulateWork()
        callServiceC(ctx)

        w.Write([]byte("Service B completed\n"))
}

func main() {
        tracer, closer, err := initTracer("serviceB")
        if err != nil {
                log.Fatalf("Could not initialize jaeger tracer: %s", err.Error())
        }
        defer closer.Close()
        opentracing.SetGlobalTracer(tracer)

        http.HandleFunc("/service-b", handler)
        log.Println("Service B started on :8081")
        log.Fatal(http.ListenAndServe(":8081", nil))
}


