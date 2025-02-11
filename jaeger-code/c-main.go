

package main

import (
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

func handler(w http.ResponseWriter, r *http.Request) {
        spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
        span := opentracing.StartSpan("serviceC", ext.RPCServerOption(spanCtx))
        defer span.Finish()

        simulateWork()

        w.Write([]byte("Service C completed\n"))
}

func main() {
        tracer, closer, err := initTracer("serviceC")
        if err != nil {
                log.Fatalf("Could not initialize jaeger tracer: %s", err.Error())
        }
        defer closer.Close()
        opentracing.SetGlobalTracer(tracer)

        http.HandleFunc("/service-c", handler)
        log.Println("Service C started on :8082")
        log.Fatal(http.ListenAndServe(":8082", nil))


}