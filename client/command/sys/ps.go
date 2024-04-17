package sys

import (
	"github.com/chainreactors/grumble"
	"github.com/chainreactors/malice-network/client/console"
	"github.com/chainreactors/malice-network/client/tui"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"
	"github.com/charmbracelet/bubbles/table"
	"google.golang.org/protobuf/proto"
	"strconv"
)

func PsCmd(ctx *grumble.Context, con *console.Console) {
	session := con.ActiveTarget.GetInteractive()
	sid := con.ActiveTarget.GetInteractive().SessionId
	if session == nil {
		return
	}
	psTask, err := con.Rpc.Ps(con.ActiveTarget.Context(), &implantpb.Request{
		Name: consts.ModulePs,
	})
	if err != nil {
		con.SessionLog(sid).Errorf("Ps error: %v", err)
		return
	}
	con.AddCallback(psTask.TaskId, func(msg proto.Message) {
		resp := msg.(*implantpb.Spite).GetPsResponse()
		var rowEntries []table.Row
		var row table.Row
		tableModel := tui.NewTable([]table.Column{
			{Title: "Name", Width: 10},
			{Title: "PID", Width: 5},
			{Title: "PPID", Width: 5},
			{Title: "Arch", Width: 7},
			{Title: "Owner", Width: 7},
			{Title: "Path", Width: 15},
			{Title: "Args", Width: 10},
		})
		for _, process := range resp.GetProcesses() {
			row = table.Row{
				process.Name,
				strconv.Itoa(int(process.Pid)),
				strconv.Itoa(int(process.Ppid)),
				process.Arch,
				process.Owner,
				process.Path,
				process.Args,
			}
			rowEntries = append(rowEntries, row)
		}
		tableModel.Rows = rowEntries
		tableModel.SetRows()
		err := tui.Run(tableModel)
		if err != nil {
			con.SessionLog(sid).Errorf("Error running table: %v", err)
		}
	})
}