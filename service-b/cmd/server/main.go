package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rzeradev/google-cloud-run/configs"
	"github.com/rzeradev/google-cloud-run/internal/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	collectorURL := os.Getenv("OTEL_EXPORTER_URL")
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(collectorURL), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("orchestrator-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func main() {
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	cfg, err := configs.LoadConfig(workdir)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	app := fiber.New()
	app.Use(logger.New())
	app.Use(otelfiber.Middleware(otelfiber.WithPropagators(propagation.TraceContext{})))
	app.Get("/weather/:zipcode", handlers.GetWeather)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
