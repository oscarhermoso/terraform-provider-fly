package providerstate

import (
	"github.com/Khan/genqlient/graphql"
)

type State struct {
	Configured    bool
	Version       string
	GraphqlClient *graphql.Client
	RestBaseUrl   string
	EnableTracing bool
	// Rest API token
	Token string
}
