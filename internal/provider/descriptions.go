package provider

import (
	"fmt"
	"regexp"
)

var (
	ID_DESC        = "A fly-generated ID"
	NAME_REGEX_RAW = `^[a-z0-9-]+$`
	NAME_REGEX     = regexp.MustCompile(NAME_REGEX_RAW)
	NAME_DESC      = fmt.Sprintf("A user-provided identifier, matching regexp `%s`", NAME_REGEX_RAW)
	APP_DESC       = "The App this resource will be created in"
	REGION_DESC    = "Fly region, ex `ord`, `sin`, `mad`"
)
