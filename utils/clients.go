package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type LoggingHttpTransport struct {
	Inner http.RoundTripper
}

func (t *LoggingHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	{
		body := []byte{}
		if req.Body != nil {
			var err error
			body, err = io.ReadAll(req.Body)
			if err != nil {
				return nil, fmt.Errorf("Error reading request body: %s", err)
			}
		}
		tflog.Debug(req.Context(), "HTTP REQUEST", map[string]any{
			"proto":   req.Proto,
			"method":  req.Method,
			"url":     req.URL.String(),
			"body":    string(body),
			"headers": req.Header,
		})
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	resp, err := t.Inner.RoundTrip(req)
	if resp != nil {
		body, err1 := io.ReadAll(resp.Body)
		if err1 != nil {
			if err != nil {
				return resp, err
			}
			return nil, fmt.Errorf("Error reading response body: %s", err1)
		}
		tflog.Debug(req.Context(), "HTTP RESPONSE", map[string]any{
			"proto":   req.Proto,
			"method":  req.Method,
			"url":     req.URL.String(),
			"code":    resp.Status,
			"body":    string(body),
			"headers": resp.Header,
		})
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return resp, err
}

type GraphqlTransport struct {
	Inner            http.RoundTripper
	Token            string
	EnableDebugTrace bool
}

func (t *GraphqlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	if t.EnableDebugTrace {
		req.Header.Add("Fly-Force-Trace", "true")
	}
	return t.Inner.RoundTrip(req)
}
