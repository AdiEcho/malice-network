package exec

import (
	"github.com/chainreactors/grumble"
	"github.com/chainreactors/malice-network/client/console"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"
	"google.golang.org/protobuf/proto"
	"os"
	"strings"
)

func ExecutePowershellCmd(ctx *grumble.Context, con *console.Console) {
	session := con.ActiveTarget.GetInteractive()
	if session == nil {
		return
	}
	sid := con.ActiveTarget.GetInteractive().SessionId
	psPath := ctx.Flags.String("path")
	var err error
	var psBin strings.Builder
	if psPath != "" {
		content, err := os.ReadFile(psPath)
		if err != nil {
			console.Log.Errorf("%s\n", err.Error())
			return
		}
		psBin.Write(content)
		psBin.WriteString("\n")
	}
	paramString := ctx.Args.StringList("args")
	psBin.WriteString(strings.Join(paramString, " "))

	task, err := con.Rpc.ExecutePowershell(con.ActiveTarget.Context(), &implantpb.ExecutePowershell{
		Name:   consts.ModuleExecutePE,
		Script: psBin.String(),
	})
	if err != nil {
		con.SessionLog(sid).Errorf("%s\n", err)
		return
	}

	con.AddCallback(task.TaskId, func(msg proto.Message) {
		resp := msg.(*implantpb.Spite)
		console.Log.Consolef("Executed Powershell on target: %s\n", resp.GetAssemblyResponse().GetData())
	})
}
