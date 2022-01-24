package http

import (
	"context"
	"strconv"
	"time"

	"github.com/releaseband/metrics/opencensus/views"

	"go.opencensus.io/stats/view"

	"github.com/releaseband/metrics/measure"
	"go.opencensus.io/tag"
)

var (
	urlKey  = tag.MustNewKey("url")
	codeKey = tag.MustNewKey("code")
)

var (
	latency       *measure.LatencyMeasure
	failedCounter *measure.CounterMeasure
)

func wrapToLatencyContext(ctx context.Context, url string) context.Context {
	if latency != nil {
		ctx, _ = tag.New(ctx, tag.Insert(urlKey, url))
	}

	return ctx
}

func record(ctx context.Context, start time.Time) {
	if latency != nil {
		latency.Record(ctx, measure.End(start))
	}
}

func commitFailedHttpCode(ctx context.Context, url string, code int) {
	if failedCounter != nil {
		ctx, _ = tag.New(ctx, tag.Insert(urlKey, url), tag.Insert(codeKey, strconv.Itoa(code)))
		failedCounter.IncrementCounter(ctx)
	}
}

func MetricViews() []*view.View {
	latency = measure.NewLatencyMeasure("http_latency", "http requests latency")

	failedCounter = measure.NewCounterMeasure("failed_http_codes", "failed http codes counter")

	return []*view.View{
		views.MakeLatencyView("http_latency", "http requests latency", latency.Measure(), []tag.Key{
			urlKey,
		}),

		views.MakeCounterView("codes", "failed http codes", failedCounter.Measure(), []tag.Key{
			urlKey,
			codeKey,
		}),
	}
}
