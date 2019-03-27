package agentDAO

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/go-kit/kit/log"
	"go.opencensus.io/trace"
	"time"
)

type LoggingMW struct {
	Logger log.Logger
	Next   Collection
}

func (mw *LoggingMW) Find(ctx context.Context, request Request) (response *mongo.Cursor, err error) {

	defer func(begin time.Time) {

		_ = mw.Logger.Log(
			"method", "Mongo Find",
			"input", spew.Sdump(request),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	response, err = mw.Next.Find(ctx, request)
	return
}

type TracingMW struct {
	Next Collection
}

func (mw *TracingMW) Find(ctx context.Context, request Request) (response *mongo.Cursor, err error) {
	ctx, span := trace.StartSpan(ctx, "Mongo Find")
	defer span.End()

	response, err = mw.Next.Find(ctx, request)
	return
}
