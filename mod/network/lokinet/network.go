package lokinet

import (
	"lokinet.io/x/mod/network"
)

// Network creates a new lokinet network driver that uses an embedded lokinet.
func Network(opts network.Opts) (network.Network, error) {
	ctx := new(EmbeddedContext)
	err := ctx.Setup(opts)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}
