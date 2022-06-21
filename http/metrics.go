package http

import (
	"context"
	_context "github.com/releaseband/metrics/v4/context"
	"go.opentelemetry.io/otel/metric/unit"
	"os"
	"time"

	"github.com/releaseband/metrics/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

const (
	successKey   = "success"
	prefixEnvKey = "RB_SERVICE"
)

var (
	successAttribute = attribute.Bool(successKey, true)
	failedAttribute  = attribute.Bool(successKey, false)
)

func getPrefix() string {
	prefix := os.Getenv(prefixEnvKey)
	if prefix == "" {
		return prefixEnvKey
	}

	return prefix
}

var (
	meter        = global.Meter(getPrefix() + ".http")
	histogram, _ = meter.SyncFloat64().Histogram(
		getPrefix()+".http_duration_seconds",
		instrument.WithDescription("http measures in seconds"),
		instrument.WithUnit(metrics.Seconds),
	)

	httpStatusCounter, _ = meter.SyncInt64().Counter(
		getPrefix()+".http_status_counter",
		instrument.WithDescription("http status counter, for statuses >= 400"),
		instrument.WithUnit(unit.Dimensionless),
	)
)

func makeHttpStatusAttribute(status int) attribute.KeyValue {
	return attribute.Int("status", status)
}

func makeProjectKey(projectKey string) attribute.KeyValue {
	return attribute.String("project_key", projectKey)
}

func record(ctx context.Context, start time.Time, isSuccess bool) {
	end := metrics.SinceInSeconds(start)
	projectID, _ := _context.GetProjectKey(ctx)

	status := successAttribute
	if !isSuccess {
		status = failedAttribute
	}

	histogram.Record(ctx, end, makeProjectKey(projectID), status)
}

func registerHttpStatus(ctx context.Context, httpCode int) {
	httpStatus := makeHttpStatusAttribute(httpCode)
	projectID, _ := _context.GetProjectKey(ctx)

	httpStatusCounter.Add(ctx, 1, makeProjectKey(projectID), httpStatus)
}
