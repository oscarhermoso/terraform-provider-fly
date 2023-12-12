package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	machineapi "github.com/andrewbaxter/terraform-provider-fly/machineapi"
	"github.com/andrewbaxter/terraform-provider-fly/providerstate"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type tfLogger struct {
	ctx context.Context
}

// Debugf implements req.Logger.
func (l *tfLogger) Debugf(format string, v ...interface{}) {
	tflog.Debug(l.ctx, fmt.Sprintf(format, v...))
}

// Errorf implements req.Logger.
func (l *tfLogger) Errorf(format string, v ...interface{}) {
	tflog.Error(l.ctx, fmt.Sprintf(format, v...))
}

// Warnf implements req.Logger.
func (l *tfLogger) Warnf(format string, v ...interface{}) {
	tflog.Warn(l.ctx, fmt.Sprintf(format, v...))
}

func NewMachineApi(ctx context.Context, state *providerstate.State) *machineapi.MachineAPI {
	out := machineapi.NewMachineAPI(state.RestBaseUrl, state.Token)
	out.HttpClient.SetLogger(&tfLogger{ctx: ctx})
	out.HttpClient.EnableDebugLog()
	if state.EnableTracing {
		out.HttpClient.SetCommonHeader("Fly-Force-Trace", "true")
		out.HttpClient.DevMode()
	}
	return out
}

type GraphqlTransport struct {
	Token            string
	EnableDebugTrace bool
}

func (t *GraphqlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	if t.EnableDebugTrace {
		req.Header.Add("Fly-Force-Trace", "true")
	}

	{
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("Error reading request body: %s", err)
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

	resp, err := http.DefaultTransport.RoundTrip(req)
	if resp != nil {
		body, err1 := io.ReadAll(req.Body)
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
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return resp, err
}
