package print

import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"kevin-rd/my-tier/internal/peer"
	"os"
)

func PrintPeers(peers []*peer.Peer) error {
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewMarkdown()))
	table.Header([]string{"ID", "VirtualIP", "RemoteAddr", "State"})
	for _, p := range peers {
		_ = table.Append([]any{p.ID, p.VirtualIP, p.RemoteAddr, p.State})
	}

	if err := table.Render(); err != nil {
		return err
	}
	return nil
}
