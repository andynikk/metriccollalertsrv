package errs

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
)

type TypeError int

type ErrorHTTP struct {
	Error error
}

type ErrorGRPC struct {
	Error error
}

var ErrStatusInternalServer = errors.New("status internal server error")
var ErrSendMsgGPRC = errors.New("send msg to GPRC error")
var ErrDecrypt = errors.New("decrypt error")
var ErrDecompress = errors.New("decompress error")
var ErrGetJSON = errors.New("get JSON error")
var ErrNotFound = errors.New("not found")
var ErrBadRequest = errors.New("bad request")
var ErrNotImplemented = errors.New("not implemented")
var ErrIPAddressAllowed = errors.New("not IP address allowed")
var ErrForbidden = errors.New("forbidden")

func StatusHTTP(e error) int {
	switch e {
	case nil:
		return http.StatusOK
	case ErrStatusInternalServer:
		return http.StatusInternalServerError
	case ErrSendMsgGPRC:
		return http.StatusInternalServerError
	case ErrDecrypt:
		return http.StatusInternalServerError
	case ErrDecompress:
		return http.StatusInternalServerError
	case ErrGetJSON:
		return http.StatusInternalServerError
	case ErrIPAddressAllowed:
		return http.StatusInternalServerError
	case ErrNotFound:
		return http.StatusNotFound
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrNotImplemented:
		return http.StatusNotImplemented
	case ErrForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func CodeGRPC(e error) codes.Code {
	switch e {
	case nil:
		return codes.OK
	case ErrStatusInternalServer:
		return codes.Internal
	case ErrSendMsgGPRC:
		return codes.Internal
	case ErrDecrypt:
		return codes.Internal
	case ErrDecompress:
		return codes.Internal
	case ErrGetJSON:
		return codes.Internal
	case ErrIPAddressAllowed:
		return codes.Internal
	case ErrNotFound:
		return codes.NotFound
	case ErrBadRequest:
		return codes.PermissionDenied
	case ErrNotImplemented:
		return codes.Unimplemented
	case ErrForbidden:
		return codes.PermissionDenied
	default:
		return codes.Internal

	}
}
