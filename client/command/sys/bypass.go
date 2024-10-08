package sys

import (
	"fmt"
	"github.com/chainreactors/malice-network/client/core"
	"github.com/chainreactors/malice-network/client/repl"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"
	"github.com/chainreactors/malice-network/proto/services/clientrpc"
	"github.com/spf13/cobra"
)

func BypassCmd(cmd *cobra.Command, con *repl.Console) {
	bypass_amsi, _ := cmd.Flags().GetBool("amsi")
	bypass_etw, _ := cmd.Flags().GetBool("etw")
	session := con.GetInteractive()
	task, err := Bypass(con.Rpc, session, bypass_amsi, bypass_etw)
	if err != nil {
		con.Log.Errorf(err.Error())
	}
	session.Console(task, fmt.Sprintf("bypass_amsi %t, bypass_etw %t", bypass_amsi, bypass_etw))
}

func Bypass(rpc clientrpc.MaliceRPCClient, session *core.Session, bypass_amsi, bypass_etw bool) (*clientpb.Task, error) {
	return rpc.Bypass(session.Context(), &implantpb.BypassRequest{
		ETW:      bypass_etw,
		AMSI:     bypass_amsi,
		BlockDll: false,
	})
}
