package connector

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"google.golang.org/grpc/codes"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
)

func getSkippEntitlementsAndGrantsAnnotations() annotations.Annotations {
	annotations := annotations.Annotations{}
	annotations.Update(&v2.SkipEntitlementsAndGrants{})

	return annotations
}

func reverseMap(m map[string]string) map[string]string {
	n := make(map[string]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func parseToUHttpError(resp *http.Response, err error) error {
	if resp == nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		// 400: Invalid JSON, malformed request, invalid parameters
		return uhttp.WrapErrorsWithRateLimitInfo(codes.InvalidArgument, resp, err)
	case http.StatusUnauthorized:
		// 401: Invalid credentials, authentication failed
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unauthenticated, resp, err)
	case http.StatusForbidden:
		// 403: Insufficient permissions
		return uhttp.WrapErrorsWithRateLimitInfo(codes.PermissionDenied, resp, err)
	case http.StatusNotFound:
		// 404: Resource not found
		return uhttp.WrapErrorsWithRateLimitInfo(codes.NotFound, resp, err)
	case http.StatusNotAcceptable:
		// 406: Invalid Accept header or API version
		return uhttp.WrapErrorsWithRateLimitInfo(codes.InvalidArgument, resp, err)
	case http.StatusRequestTimeout:
		// 408: Request timeout
		return uhttp.WrapErrorsWithRateLimitInfo(codes.DeadlineExceeded, resp, err)
	case http.StatusConflict:
		// 409: Resource already exists, duplicate
		return uhttp.WrapErrorsWithRateLimitInfo(codes.AlreadyExists, resp, err)
	case http.StatusUnprocessableEntity:
		// 422: Business logic validation failed
		return uhttp.WrapErrorsWithRateLimitInfo(codes.FailedPrecondition, resp, err)
	case http.StatusTooManyRequests:
		// 429: Rate limit exceeded
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, err)
	case http.StatusNotImplemented:
		// 501: Feature not implemented
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unimplemented, resp, err)
	case http.StatusServiceUnavailable:
		// 503: Service unavailable
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, err)
	}

	// 5xx: Server errors
	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, err)
	}

	// Other non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unknown, resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	return errors.Join(
		fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		err,
	)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
