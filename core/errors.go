package core

import "fmt"

type errInvalidEndpoint struct {
	rpc string
}

func (e errInvalidEndpoint) Error() string {
	return fmt.Sprintf("invalid rpc endpoint fmt: %s, it must be HTTP or HTTPS protocol", e.rpc)
}
