package runtime

import ferretnet "github.com/MontFerret/ferret/v2/pkg/net"

type networkCleanup struct {
	ferretnet.Network
	close func()
}

func (n *networkCleanup) CloseIdleConnections() {
	n.close()
}
