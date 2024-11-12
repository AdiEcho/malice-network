package generic

import (
	"fmt"
	"github.com/chainreactors/malice-network/client/command/common"
	"github.com/chainreactors/malice-network/client/core"
	"github.com/chainreactors/malice-network/client/core/intermediate"
	"github.com/chainreactors/malice-network/client/repl"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/helper/proto/client/clientpb"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"os/exec"
	"strings"
)

func Commands(con *repl.Console) []*cobra.Command {
	loginCmd := &cobra.Command{
		Use:   consts.CommandLogin,
		Short: "Login to server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return LoginCmd(cmd, con)
		},
	}

	versionCmd := &cobra.Command{
		Use:   consts.CommandVersion,
		Short: "show server version",
		Run: func(cmd *cobra.Command, args []string) {
			VersionCmd(cmd, con)
			return
		},
	}

	exitCmd := &cobra.Command{
		Use:   consts.CommandExit,
		Short: "exit client",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Exiting...")
			os.Exit(0)
			return
		},
	}

	broadcastCmd := &cobra.Command{
		Use:   consts.CommandBroadcast + " [message]",
		Short: "Broadcast a message to all clients",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			BroadcastCmd(cmd, con)
		},
	}

	common.BindFlag(broadcastCmd, func(f *pflag.FlagSet) {
		f.BoolP("notify", "n", false, "notify the message to third-party services")
	})

	cmdCmd := &cobra.Command{
		Use:   "! [command]",
		Short: "Run a command",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// os exec

			out, err := exec.Command(args[0], args[1:]...).Output()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// 打印标准输出
			fmt.Println(string(out))
		},
	}

	return []*cobra.Command{loginCmd, versionCmd, exitCmd, broadcastCmd, cmdCmd}
}

func Log(con *repl.Console, sess *core.Session, msg string, notify bool) (bool, error) {
	_, err := con.Rpc.SessionEvent(sess.Context(), &clientpb.Event{
		Type:    consts.EventSession,
		Op:      consts.CtrlSessionLog,
		Session: sess.Session,
		Client:  con.Client,
		Message: []byte(msg),
	})
	if err != nil {
		return false, err
	}
	if notify {
		return Notify(con, &clientpb.Event{
			Type:    consts.EventNotify,
			Client:  con.Client,
			Message: []byte(msg),
		})
	}
	return true, nil
}

func Register(con *repl.Console) {
	con.RegisterServerFunc(consts.CommandBroadcast, func(con *repl.Console, msg string) (bool, error) {
		return Broadcast(con, &clientpb.Event{
			Type:    consts.EventBroadcast,
			Client:  con.Client,
			Message: []byte(msg),
		})
	}, nil)

	con.RegisterServerFunc(consts.CommandNotify, func(con *repl.Console, msg string) (bool, error) {
		return Notify(con, &clientpb.Event{
			Type:    consts.EventNotify,
			Client:  con.Client,
			Message: []byte(msg),
		})
	}, nil)

	con.RegisterServerFunc("callback_log", func(con *repl.Console, sess *core.Session, notify bool) intermediate.BuiltinCallback {
		return func(content interface{}) (bool, error) {
			return Log(con, sess, fmt.Sprintf("%v", content), notify)
		}
	}, nil)

	con.RegisterServerFunc("log", func(con *repl.Console, sess *core.Session, msg string, notify bool) (bool, error) {
		return Log(con, sess, msg, notify)
	}, nil)

	con.RegisterServerFunc("blog", func(con *repl.Console, sess *core.Session, msg string) (bool, error) {
		return Log(con, sess, msg, false)
	}, nil)

	con.RegisterServerFunc("barch", func(con *repl.Console, sess *core.Session) (string, error) {
		return sess.Os.Arch, nil
	}, nil)

	con.RegisterServerFunc("active", func(con *repl.Console) (*core.Session, error) {
		return con.GetInteractive().Clone(consts.CalleeMal), nil
	}, &intermediate.Helper{
		Short:   "get current session",
		Output:  []string{"sess"},
		Example: "active()",
	})

	con.RegisterServerFunc("is64", func(con *repl.Console, sess *core.Session) (bool, error) {
		return sess.Os.Arch == "x64", nil
	}, nil)

	con.RegisterServerFunc("isactive", func(con *repl.Console, sess *core.Session) (bool, error) {
		return sess.IsAlive, nil
	}, nil)

	con.RegisterServerFunc("isadmin", func(con *repl.Console, sess *core.Session) (bool, error) {
		return sess.IsPrivilege, nil
	}, nil)

	con.RegisterServerFunc("isbeacon", func(con *repl.Console, sess *core.Session) (bool, error) {
		return sess.Type == consts.ImplantTypeBeacon, nil
	}, nil)

	con.RegisterServerFunc("donut_exe2shellcode", func(exe []byte, arch string, param string) (string, error) {
		cmdline, err := shellquote.Split(param)
		if err != nil {
			return "", err
		}

		bin, err := con.Rpc.EXE2Shellcode(con.Context(), &clientpb.EXE2Shellcode{
			Bin:    exe,
			Arch:   arch,
			Type:   "donut",
			Params: strings.Join(cmdline, ","),
		})
		if err != nil {
			return "", err
		}
		return string(bin.Bin), nil
	}, nil)

	con.RegisterServerFunc("donut_dll2shellcode", func(dll []byte, arch string, param string) (string, error) {
		cmdline, err := shellquote.Split(param)
		if err != nil {
			return "", err
		}

		bin, err := con.Rpc.DLL2Shellcode(con.Context(), &clientpb.DLL2Shellcode{
			Bin:    dll,
			Arch:   arch,
			Type:   "donut",
			Params: strings.Join(cmdline, ","),
		})
		if err != nil {
			return "", err
		}
		return string(bin.Bin), nil
	}, nil)

	con.RegisterServerFunc("srdi", func(dll []byte, entry string, arch string, param string) (string, error) {
		bin, err := con.Rpc.DLL2Shellcode(con.Context(), &clientpb.DLL2Shellcode{
			Bin:        dll,
			Arch:       arch,
			Type:       "srdi",
			Entrypoint: entry,
			Params:     param,
		})
		if err != nil {
			return "", err
		}
		return string(bin.Bin), nil
	}, nil)

	con.RegisterServerFunc("sgn_encode", func(shellcode []byte, arch string, iterations int32) (string, error) {
		bin, err := con.Rpc.ShellcodeEncode(con.Context(), &clientpb.ShellcodeEncode{
			Shellcode:  shellcode,
			Arch:       arch,
			Type:       "sgn",
			Iterations: iterations,
		})
		if err != nil {
			return "", err
		}
		return string(bin.Bin), nil
	}, nil)
}
