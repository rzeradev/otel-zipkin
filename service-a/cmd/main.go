package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"unicode"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

type CustomTransport struct {
	base http.RoundTripper
}

func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Extract the current span
	span := trace.SpanFromContext(req.Context())
	if span.SpanContext().IsValid() {
		// Set the span name based on the HTTP request details
		span.SetName(req.Method + " " + req.URL.Path)
	}
	return t.base.RoundTrip(req)
}

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
			semconv.ServiceNameKey.String("input-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{}))
	return tp, nil
}

func main() {
	app := fiber.New()
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	app.Use(logger.New())
	app.Use(otelfiber.Middleware(
		otelfiber.WithSpanNameFormatter(func(ctx *fiber.Ctx) string {
			return ctx.Method() + " " + ctx.Path()
		}),
	))

	client := &http.Client{
		Transport: otelhttp.NewTransport(&CustomTransport{base: http.DefaultTransport}),
	}

	app.Post("/weather", func(c *fiber.Ctx) error {
		var request struct {
			CEP string `json:"cep"`
		}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "invalid input"})
		}

		if len(request.CEP) != 8 {
			return c.Status(422).JSON(fiber.Map{"message": "invalid zipcode"})
		}

		if !IsNumeric(request.CEP) {
			return c.Status(422).JSON(fiber.Map{"message": "invalid zipcode"})
		}

		req, err := http.NewRequestWithContext(c.UserContext(), "GET", os.Getenv("ORCHESTRATOR_URL")+"/weather/"+request.CEP, nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "internal server error"})
		}

		resp, err := client.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "internal server error"})
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "internal server error"})
		}

		return c.Status(resp.StatusCode).JSON(result)
	})

	log.Fatal(app.Listen(":8080"))
}
