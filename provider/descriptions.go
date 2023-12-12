package provider

import (
	"fmt"
	"regexp"
)

var (
	ID_DESC                = "A fly-generated ID"
	NAME_REGEX_RAW         = `^[a-z0-9-]+$`
	NAME_REGEX             = regexp.MustCompile(NAME_REGEX_RAW)
	NAME_DESC              = fmt.Sprintf("A user-provided identifier, matching regexp `%s`", NAME_REGEX_RAW)
	APP_DESC               = "The App this resource will be created in"
	REGION_DESC            = "Fly region, ex `ord`, `sin`, `mad`"
	SHAREDIP_DESC          = "A shared ipv4 address, automatically attached in certain conditions or if explicitly requested"
	ADDRESS_TYPE_REGEX_RAW = `^(v4|v6|private_v6)$`
	ADDRESS_TYPE_REGEX     = regexp.MustCompile(ADDRESS_TYPE_REGEX_RAW)
	ADDRESS_TYPE_DESC      = fmt.Sprintf("One of the following values (by regex): `%s`", ADDRESS_TYPE_REGEX_RAW)
)
