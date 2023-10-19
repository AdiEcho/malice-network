package exec

import (
	"context"
	"fmt"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/malice-network/client/console"
	"github.com/chainreactors/malice-network/proto/implant/pluginpb"
	"github.com/desertbit/grumble"
)

func ExecuteCmd(ctx *grumble.Context, con *console.Console) {
	session := con.ActiveTarget.GetInteractive()
	if session == nil {
		return
	}

	cmdPath := ctx.Args.String("command")
	args := ctx.Args.StringList("arguments")
	//token := ctx.Flags.Bool("token")
	output := ctx.Flags.Bool("output")
	stdout := ctx.Flags.String("stdout")
	stderr := ctx.Flags.String("stderr")
	saveLoot := ctx.Flags.Bool("loot")
	saveOutput := ctx.Flags.Bool("save")
	//ppid := ctx.Flags.Uint("ppid")
	//hostName := getHostname(session, beacon)
	var captureOutput bool = output || saveLoot || saveOutput

	if output {
		logs.Log.Error("Using --output in beacon mode, if the command blocks the task will never complete\n")
	}

	var exec *pluginpb.ExecResponse
	var err error

	ctrl := make(chan bool)
	//con.SpinUntil(fmt.Sprintf("Executing %s %s ...", cmdPath, strings.Join(args, " ")), ctrl)
	exec, err = con.Rpc.Execute(context.Background(), &pluginpb.ExecRequest{
		Path:   cmdPath,
		Args:   args,
		Output: captureOutput,
		Stderr: stderr,
		Stdout: stdout,
	})
	ctrl <- true
	<-ctrl
	if err != nil {
		logs.Log.Errorf("%s", err.Error())
		return
	}

	fmt.Println(exec)
}
