package middlware

import (
	"context"
	"net/http"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/networks"
)

type KeyValueContext string

func CheckIP(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		xRealIP := r.Header.Get("X-Real-IP")
		if xRealIP == "" {
			w.WriteHeader(http.StatusOK)
			endpoint(w, r)
			return
		}

		ok := networks.AddressAllowed(strings.Split(xRealIP, constants.SepIPAddress))
		if ok {
			w.WriteHeader(http.StatusOK)
			endpoint(w, r)
			return
		}

		w.WriteHeader(http.StatusForbidden)
		_, err := w.Write([]byte("Not IP address allowed"))
		if err != nil {
			constants.Logger.ErrorLog(err)
		}
	})
}

func ChiCheckIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		key := KeyValueContext("IP-Address-Allowed")

		xRealIP := r.Header.Get("X-Real-IP")
		ctx := context.WithValue(r.Context(), key, "false")
		if xRealIP == "" {
			ctx = context.WithValue(r.Context(), key, "true")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ok := networks.AddressAllowed(strings.Split(xRealIP, constants.SepIPAddress))
		if ok {
			ctx = context.WithValue(r.Context(), key, "true")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//func serverInterceptor(ctx context.Context,
//	req interface{},
//	info *grpc.UnaryServerInfo,
//	handler grpc.UnaryHandler) (interface{}, error) {
//
//	md, ok := metadata.FromIncomingContext(ctx)
//	if !ok {
//		return nil, nil
//	}
//	xRealIP := md[strings.ToLower("X-Real-IP")]
//	for _, val := range xRealIP {
//		ok = networks.AddressAllowed(strings.Split(val, constants.SepIPAddress))
//		if !ok {
//			return nil, nil
//		}
//	}
//	h, _ := handler(ctx, req)
//	return h, nil
//}
//
//func WithServerUnaryInterceptor() grpc.ServerOption {
//	return grpc.UnaryInterceptor(serverInterceptor)
//}