package main

import (
	"context"
	"log"

	ddmetrics "github.com/raywall/dd-metrics"
)

var (
	handler *ddmetrics.Handler
	err     error
)

func init() {
	tags := []string{
		"env:development",
		"service:example_app",
		"version:1.0.0",
	}

	handler, err = ddmetrics.NewHandler(context.Background(), ddmetrics.JSON, tags)
	if err != nil {
		log.Fatal("Erro ao criar handler:", err)
	}

	err = handler.StartTrace("app-tester", "domain", "service", "127.0.0.1", "8126")
	if err != nil {
		log.Fatal("Erro ao iniciar trace:", err)
	}

	err = handler.StartMetric("127.0.0.1", "8125", "custom.")
	if err != nil {
		log.Fatal("Erro ao iniciar métrica:", err)
	}
}

func main() {
	defer handler.StopTrace()
	defer handler.StopMetric()

	handler.SendMetric("example.metric", 42.0)
	log.Println("Trace sent to OpenTelemetry Collector")
}
