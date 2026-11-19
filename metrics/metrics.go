package metrics

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

var (
	commandExecutions otelMetric.Int64Counter
	parserExecutions  otelMetric.Int64Counter
	receivedMessages  otelMetric.Int64Counter
	parserDuration    otelMetric.Float64Histogram
	commandDuration   otelMetric.Float64Histogram
	meterProvider     *metric.MeterProvider
	meter             = otel.Meter("regular-bot")
)

func CommandInc(ctx context.Context, command string, succeded bool) {
	commandExecutions.Add(ctx, 1,
		otelMetric.WithAttributes(
			attribute.KeyValue{
				Key:   "name",
				Value: attribute.StringValue(command),
			},
			attribute.KeyValue{
				Key:   "succeded",
				Value: attribute.BoolValue(succeded),
			},
		),
	)
}

func ParserInc(ctx context.Context, parser string, succeded bool) {
	parserExecutions.Add(ctx, 1,
		otelMetric.WithAttributes(
			attribute.KeyValue{
				Key:   "name",
				Value: attribute.StringValue(parser),
			},
			attribute.KeyValue{
				Key:   "succeded",
				Value: attribute.BoolValue(succeded),
			},
		),
	)
}

func ParserDuration(ctx context.Context, parser string, d time.Duration) {
	parserDuration.Record(ctx, d.Seconds(),
		otelMetric.WithAttributes(
			attribute.KeyValue{
				Key:   "name",
				Value: attribute.StringValue(parser),
			},
		),
	)
}

func CommandDuration(ctx context.Context, command string, d time.Duration) {
	commandDuration.Record(ctx, d.Seconds(),
		otelMetric.WithAttributes(
			attribute.KeyValue{
				Key:   "name",
				Value: attribute.StringValue(command),
			},
		),
	)
}

func MessagesInc(ctx context.Context) {
	receivedMessages.Add(ctx, 1)
}

func Setup() error {
	ctx := context.Background()
	exp, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return err
	}

	meterProvider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(meterProvider)

	receivedMessages, err = meter.Int64Counter(
		"message.processed",
		otelMetric.WithDescription("Number of processed messages"),
		// otelMetric.WithAttributes()
		otelMetric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	commandExecutions, err = meter.Int64Counter(
		"command.executed",
		otelMetric.WithDescription("Number of executed commands"),
		// otelMetric.WithAttributes()
		otelMetric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	commandDuration, err = meter.Float64Histogram(
		"command.duration",
		otelMetric.WithDescription("The duration of command execution."),
		otelMetric.WithUnit("s"),
		otelMetric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10),
	)
	if err != nil {
		return err
	}

	parserExecutions, err = meter.Int64Counter(
		"parser.executed",
		otelMetric.WithDescription("Number of executed parsers"),
		// otelMetric.WithAttributes(),
		otelMetric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	parserDuration, err = meter.Float64Histogram(
		"parser.duration",
		otelMetric.WithDescription("The duration of parser execution."),
		otelMetric.WithUnit("s"),
		otelMetric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10),
	)
	if err != nil {
		return err
	}

	return nil
}

func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := meterProvider.Shutdown(ctx); err != nil {
		slog.Error("couldn't stop otlp", "err", err)
	}
}
