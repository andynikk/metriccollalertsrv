package middlware

import (
	"context"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil
	}
	xRealIP := md[strings.ToLower("X-Real-IP")]
	addressAllowed := md[strings.ToLower("IP-Address-Allowed")]

	for _, val := range xRealIP {
		for _, valA := range addressAllowed {
			ok = networks.AddressAllowed(strings.Split(val, constants.SepIPAddress), valA)
			if !ok {
				return nil, nil
			}
		}
	}
	h, _ := handler(ctx, req)
	return h, nil
}

func WithServerUnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(serverInterceptor)
}
