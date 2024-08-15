package exec

import (
	"github.com/chainreactors/grumble"
	"github.com/chainreactors/malice-network/client/console"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"

	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
)

func ExecuteBofCmd(ctx *grumble.Context, con *console.Console) {
	session := con.GetInteractive()
	if session == nil {
		return
	}
	path := ctx.Args.String("path")
	args := ctx.Args.StringList("args")
	name := filepath.Base(path)
	binData, err := os.ReadFile(path)
	if err != nil {
		console.Log.Errorf("%s\n", err)
		return
	}

	var task *clientpb.Task
	task, err = con.Rpc.ExecuteBof(con.ActiveTarget.Context(), &implantpb.ExecuteBinary{
		Name:   name,
		Bin:    binData,
		Params: args,
		Output: true,
		Type:   consts.ModuleExecuteBof,
	})

	if err != nil {
		console.Log.Errorf("%s", err.Error())
		return
	}

	con.AddCallback(task.TaskId, func(msg proto.Message) {
		resp := msg.(*implantpb.Spite).GetAssemblyResponse()

		con.SessionLog(con.GetInteractive().SessionId).Infof("%s output:\n%s", name, string(resp.Data))
	})
}
