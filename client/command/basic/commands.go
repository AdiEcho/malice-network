package basic

import (
	"github.com/chainreactors/malice-network/client/command/common"
	"github.com/chainreactors/malice-network/client/core"
	"github.com/chainreactors/malice-network/client/repl"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/helper/proto/client/clientpb"
	"github.com/chainreactors/malice-network/helper/proto/services/clientrpc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Commands(con *repl.Console) []*cobra.Command {
	sleepCmd := &cobra.Command{
		Use:   consts.ModuleSleep + " [interval/second]",
		Short: "change implant sleep config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return SleepCmd(cmd, con)
		},
	}

	common.BindFlag(sleepCmd, func(f *pflag.FlagSet) {
		f.Float64("jitter", 0, "jitter")
	})

	suicideCmd := &cobra.Command{
		Use:   consts.ModuleSuicide,
		Short: "kill implant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return SuicideCmd(cmd, con)
		},
	}

	pingCmd := &cobra.Command{
		Use:   consts.ModulePing,
		Short: "check if implant is alive",
		RunE: func(cmd *cobra.Command, args []string) error {
			return PingCmd(cmd, con)
		},
	}
	return []*cobra.Command{sleepCmd, suicideCmd, pingCmd}
}

func Register(con *repl.Console) {
	con.RegisterImplantFunc(consts.ModulePing, Ping, "", nil, common.ParseStatus, nil)

	con.RegisterImplantFunc(consts.ModuleSleep,
		Sleep,
		"bsleep",
		func(rpc clientrpc.MaliceRPCClient, sess *core.Session, interval uint64) (*clientpb.Task, error) {
			return Sleep(rpc, sess, interval, sess.Timer.Jitter)
		},
		common.ParseStatus,
		nil,
	)

	con.AddInternalFuncHelper(consts.ModuleSleep, consts.ModuleSleep,
		`sleep(active(), 10, 0.5)`,
		[]string{
			"sess:special session",
			"interval:time interval, in seconds",
			"jitter:jitter, percentage of interval",
		}, []string{"task"})

	con.RegisterImplantFunc(consts.ModuleSuicide,
		Suicide,
		"bexit",
		nil,
		common.ParseStatus,
		nil,
	)

	con.AddInternalFuncHelper(consts.ModuleSuicide, consts.ModuleSuicide,
		`suicide(active())`,
		[]string{
			"sess:special session",
		},
		[]string{"task"},
	)
}
