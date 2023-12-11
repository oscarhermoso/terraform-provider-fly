package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/fly-apps/terraform-provider-fly/internal/providerstate"
	"github.com/fly-apps/terraform-provider-fly/pkg/apiv1"
	hreq "github.com/imroc/req/v3"

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

func NewMachineApi(ctx context.Context, state *providerstate.State) *apiv1.MachineAPI {
	httpClient := hreq.C()
	httpClient.SetLogger(&tfLogger{ctx: ctx})
	httpClient.EnableDebugLog()

	if state.EnableTracing {
		httpClient.SetCommonHeader("Fly-Force-Trace", "true")
		httpClient = hreq.C().DevMode()
	}

	httpClient.SetCommonHeader("Authorization", "Bearer "+state.Token)
	httpClient.SetTimeout(2 * time.Minute)
	return apiv1.NewMachineAPI(httpClient, state.RestBaseUrl)
}
