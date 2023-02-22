package utils

import (
	"context"
	"fmt"

	"github.com/fly-apps/terraform-provider-fly/internal/providerstate"
	"github.com/fly-apps/terraform-provider-fly/pkg/apiv1"

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
	out := apiv1.NewMachineAPI(state.RestBaseUrl, state.Token)
	out.HttpClient.SetLogger(&tfLogger{ctx: ctx})
	out.HttpClient.EnableDebugLog()
	if state.EnableTracing {
		out.HttpClient.SetCommonHeader("Fly-Force-Trace", "true")
		out.HttpClient.DevMode()
	}
	return out
}
