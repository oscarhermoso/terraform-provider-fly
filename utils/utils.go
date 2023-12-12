package utils

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
