package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func P[T any](x T) *T {
	return &x
}

func HandleGraphqlErrors(diagnostics *diag.Diagnostics, err error, format string, formatArgs ...any) {
	var errList gqlerror.List
	if errors.As(err, &errList) {
		for _, err := range errList {
			diagnostics.AddError(
				fmt.Sprintf(format, formatArgs...),
				fmt.Sprintf("Upstream graphql error: %s\nGraphql path: %s", err.Message, err.Path.String()),
			)
		}
	} else if err != nil {
		diagnostics.AddError(
			fmt.Sprintf(format, formatArgs...),
			fmt.Sprintf("Upstream graphql error: %s", err.Error()),
		)
	}
}

// Retry runs the function repeatedly until it succeeds, it fails with `finalErr`, or until the specified total duration
// elapses. Between each run it sleeps for the `period` time. It returns the last error upon failure, and logs all errors
// as it runs at `DEBUG` level.
func Retry(ctx context.Context, totalTime time.Duration, period time.Duration, handler func() (temporaryErr error, finalErr error)) error {
	endTime := time.Now().Add(totalTime)
	count := 0
	var lastErr error = nil
	for {
		if count >= 2 {
			if time.Now().After(endTime) {
				return fmt.Errorf(
					"All %d attempts over %s failed before retry timeout; last error: %s",
					count,
					(time.Now().Sub(endTime) + totalTime).Truncate(time.Second),
					lastErr,
				)
			}
		}

		// Run
		count += 1
		tempErr, finalErr := handler()
		if finalErr == nil && tempErr == nil {
			return nil
		}

		// React to errors
		if finalErr != nil {
			return finalErr
		} else if tempErr != nil {
			tflog.Debug(context.TODO(), fmt.Sprintf("Retry failed with temporary error: %#v", tempErr))
			lastErr = tempErr
		} else {
			panic("UNREACHABLE")
		}

		// Sleep before next iteration
		slept := make(chan bool, 1)
		go func() {
			time.Sleep(period)
			slept <- true
		}()
		select {
		case <-slept:
			// nop
		case <-ctx.Done():
			return fmt.Errorf("Cancelled")
		}
	}
}
