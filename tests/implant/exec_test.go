package implant

import (
	"fmt"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/helper/encoders/hash"
	"github.com/chainreactors/malice-network/helper/types"
	"github.com/chainreactors/malice-network/proto/implant/commonpb"
	"github.com/chainreactors/malice-network/proto/implant/pluginpb"
	"github.com/chainreactors/malice-network/tests/common"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	implant := common.NewImplant(common.DefaultListenerAddr, common.TestSid)
	implant.Register()
	time.Sleep(1 * time.Second)
	rpc := common.NewClient(common.DefaultGRPCAddr, common.TestSid)
	fmt.Println(hash.Md5Hash([]byte(implant.Sid)))
	go func() {
		conn := implant.MustConnect()
		implant.WriteEmpty(conn)
		res, err := implant.Read(conn)
		fmt.Printf("res %v %v\n", res, err)
		spite := &commonpb.Spite{
			TaskId: 0,
			End:    true,
		}
		resp := &pluginpb.ExecResponse{
			Stdout:     []byte("admin"),
			Pid:        999,
			StatusCode: 0,
		}
		types.BuildSpite(spite, resp)
		err = implant.WriteSpite(conn, spite)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	time.Sleep(1 * time.Second)
	exec := &pluginpb.ExecRequest{
		Path: "/bin/bash",
		Args: []string{"whoami"},
	}
	resp, err := rpc.Call(consts.ExecutionStr, exec)
	if err != nil {
		return
	}
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("resp %v\n", resp)

}
