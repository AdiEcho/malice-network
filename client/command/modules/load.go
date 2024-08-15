package modules

import (
	"github.com/chainreactors/grumble"
	"github.com/chainreactors/malice-network/client/console"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"
	"google.golang.org/protobuf/proto"
	"os"
)

func loadModule(ctx *grumble.Context, con *console.Console) {
	session := con.GetInteractive()
	if session == nil {
		return
	}
	sid := con.GetInteractive().SessionId
	bundle := ctx.Flags.String("name")
	path := ctx.Args.String("path")
	data, err := os.ReadFile(path)
	if err != nil {
		console.Log.Errorf("Error reading file: %v", err)
		return
	}
	loadTask, err := con.Rpc.LoadModule(con.ActiveTarget.Context(), &implantpb.LoadModule{
		Bundle: bundle,
		Bin:    data,
	})
	if err != nil {
		console.Log.Errorf("LoadModule error: %v", err)
		return
	}
	con.AddCallback(loadTask.TaskId, func(msg proto.Message) {
		//modules := msg.(*implantpb.Spite).GetModules()
		con.SessionLog(sid).Infof("LoadModule: success")
	})
}
