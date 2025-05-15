package print

import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"os"
)

func Print(peers any) error {
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewMarkdown()))
	table.Header([]string{"ID", "CIDR", "RemoteAddr", "State"})
	if err := table.Bulk(peers); err != nil {
		return err
	}
	if err := table.Render(); err != nil {
		return err
	}
	return nil
}
