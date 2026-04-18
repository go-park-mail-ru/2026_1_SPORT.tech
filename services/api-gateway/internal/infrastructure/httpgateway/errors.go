package httpgateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	httpMetadataStatusCodeKey  = "x-http-status-code"
	httpMetadataSetCookieKey   = "x-http-set-cookie"
	httpMetadataClearCookieKey = "x-http-clear-cookie"
	httpMetadataHeaderKey      = "x-http-header"
)

func forwardResponseOption(ctx context.Context, writer http.ResponseWriter, _ proto.Message) error {
	serverMetadata, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	for _, cookie := range serverMetadata.HeaderMD.Get(httpMetadataSetCookieKey) {
		writer.Header().Add("Set-Cookie", cookie)
	}
	for _, cookie := range serverMetadata.HeaderMD.Get(httpMetadataClearCookieKey) {
		writer.Header().Add("Set-Cookie", cookie)
	}
	for _, headerValue := range serverMetadata.HeaderMD.Get(httpMetadataHeaderKey) {
		parts := strings.SplitN(headerValue, ":", 2)
		if len(parts) != 2 {
			continue
		}
		writer.Header().Set(parts[0], parts[1])
	}

	statusCodes := serverMetadata.HeaderMD.Get(httpMetadataStatusCodeKey)
	if len(statusCodes) == 0 {
		return nil
	}

	httpStatusCode, err := strconv.Atoi(statusCodes[0])
	if err != nil {
		return nil
	}

	if httpStatusCode == http.StatusNoContent {
		writer.Header().Del("Content-Type")
	}
	writer.WriteHeader(httpStatusCode)

	return nil
}

func httpErrorHandler(
	_ context.Context,
	_ *runtime.ServeMux,
	_ runtime.Marshaler,
	writer http.ResponseWriter,
	_ *http.Request,
	err error,
) {
	st := status.Convert(err)
	httpStatusCode := runtime.HTTPStatusFromCode(st.Code())
	writePublicError(writer, httpStatusCode, publicErrorCode(st.Code()), st.Message())
}

func routingErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	writer http.ResponseWriter,
	request *http.Request,
	httpStatusCode int,
) {
	switch httpStatusCode {
	case http.StatusNotFound:
		httpErrorHandler(ctx, mux, marshaler, writer, request, status.Error(codes.NotFound, http.StatusText(httpStatusCode)))
	case http.StatusMethodNotAllowed:
		httpErrorHandler(ctx, mux, marshaler, writer, request, status.Error(codes.Unimplemented, http.StatusText(httpStatusCode)))
	case http.StatusBadRequest:
		httpErrorHandler(ctx, mux, marshaler, writer, request, status.Error(codes.InvalidArgument, http.StatusText(httpStatusCode)))
	default:
		httpErrorHandler(ctx, mux, marshaler, writer, request, status.Error(codes.Internal, http.StatusText(httpStatusCode)))
	}
}

func publicErrorCode(code codes.Code) string {
	switch code {
	case codes.InvalidArgument:
		return "bad_request"
	case codes.AlreadyExists, codes.Aborted:
		return "already_exists"
	case codes.Unauthenticated:
		return "unauthorized"
	case codes.PermissionDenied:
		return "forbidden"
	case codes.NotFound:
		return "not_found"
	case codes.Unimplemented:
		return "not_implemented"
	case codes.Internal, codes.Unknown:
		return "internal_error"
	default:
		return strings.ToLower(code.String())
	}
}

func writeGRPCError(writer http.ResponseWriter, err error) {
	st := status.Convert(err)
	writePublicError(writer, runtime.HTTPStatusFromCode(st.Code()), publicErrorCode(st.Code()), st.Message())
}

func writePublicError(writer http.ResponseWriter, httpStatusCode int, code string, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(httpStatusCode)

	_ = json.NewEncoder(writer).Encode(&gatewayv1.ErrorResponse{
		Error: &gatewayv1.Error{
			Code:    code,
			Message: message,
		},
	})
}
