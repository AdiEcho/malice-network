package rpc

import (
	"context"
	"github.com/chainreactors/malice-network/proto/implant/pluginpb"
)

func (rpc *Server) Exec(ctx context.Context, req *pluginpb.ExecRequest) (*pluginpb.ExecResponse, error) {
	resp := &pluginpb.ExecResponse{}
	err := rpc.GenericHandler(ctx, req, resp)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
