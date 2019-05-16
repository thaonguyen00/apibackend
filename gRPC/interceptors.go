package gRPC

import (
	"github.com/davecgh/go-spew/spew"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"os"
	"time"
	"golang.org/x/net/context"
	"fmt"
	klog "github.com/go-kit/kit/log"
)

func GrpcTracer(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	ctx, span := trace.StartSpan(ctx, "GRPC Request")
	defer span.End()
	span.Annotate([]trace.Attribute{
		trace.StringAttribute("request", fmt.Sprintf("%v", req)),
	}, "Input")

	return handler(ctx, req)
}

func GrpcLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := klog.NewLogfmtLogger(os.Stderr)
	defer func(begin time.Time) {
		_ = logger.Log(
			"method", "GRPC Request",
			"err", err,
			"request", spew.Sdump(req),
			"took", time.Since(begin),
		)
	}(time.Now())

	return handler(ctx, req)
}


