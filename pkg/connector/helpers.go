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

func wrapError(err error, message string) error {
	return fmt.Errorf("mongo-db-connector: %s: %w", message, err)
}

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
	case http.StatusRequestTimeout:
		return uhttp.WrapErrorsWithRateLimitInfo(codes.DeadlineExceeded, resp, err)
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, err)
	case http.StatusNotFound:
		return uhttp.WrapErrorsWithRateLimitInfo(codes.NotFound, resp, err)
	case http.StatusUnauthorized:
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unauthenticated, resp, err)
	case http.StatusForbidden:
		return uhttp.WrapErrorsWithRateLimitInfo(codes.PermissionDenied, resp, err)
	case http.StatusNotImplemented:
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unimplemented, resp, err)
	}

	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unavailable, resp, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return uhttp.WrapErrorsWithRateLimitInfo(codes.Unknown, resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	return errors.Join(
		fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		err,
	)
}
