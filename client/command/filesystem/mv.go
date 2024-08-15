package filesystem

import (
	"github.com/chainreactors/grumble"
	"github.com/chainreactors/malice-network/client/console"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"
	"google.golang.org/protobuf/proto"
)

func MvCmd(ctx *grumble.Context, con *console.Console) {
	session := con.GetInteractive()
	if session == nil {
		return
	}
	sid := con.GetInteractive().SessionId
	sourcePath := ctx.Flags.String("source")
	targetPath := ctx.Flags.String("target")
	args := []string{sourcePath, targetPath}
	mvTask, err := con.Rpc.Mv(con.ActiveTarget.Context(), &implantpb.Request{
		Name: consts.ModuleMv,
		Args: args,
	})
	if err != nil {
		console.Log.Errorf("Mv error: %v", err)
		return
	}
	con.AddCallback(mvTask.TaskId, func(msg proto.Message) {
		_ = msg.(*implantpb.Spite)
		con.SessionLog(sid).Consolef("Mv success\n")
	})
}
