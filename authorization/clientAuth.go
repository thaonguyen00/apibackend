package authorization

import (
	"google.golang.org/grpc/metadata"
	"context"
	"fmt"
)

func AddAuth(ctx context.Context, token string) context.Context {
	auth := fmt.Sprintf("Bearer %s", token)
	ctx = metadata.AppendToOutgoingContext(ctx, "grpcgateway-authorization", auth)
	return ctx
}
