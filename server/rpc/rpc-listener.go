package rpc

import (
	"context"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/commonpb"
	"github.com/chainreactors/malice-network/proto/listener/lispb"
	"github.com/chainreactors/malice-network/proto/services/listenerrpc"
	"github.com/chainreactors/malice-network/server/core"
	"google.golang.org/grpc/peer"
)

func (rpc *Server) GetListeners(ctx context.Context, req *clientpb.Empty) (*clientpb.Listeners, error) {
	return core.Listeners.ToProtobuf(), nil
}

func (rpc *Server) RegisterListener(ctx context.Context, req *lispb.RegisterListener) (*commonpb.Empty, error) {
	core.Listeners.Add(&core.Listener{
		Name:   req.Name,
		Host:   req.Addr,
		Active: true,
	})
	p, ok := peer.FromContext(ctx)
	if !ok {
		return &commonpb.Empty{}, nil
	}
	logs.Log.Importantf("%s register listener %s", p.Addr, req.Name)
	return &commonpb.Empty{}, nil
}

func (rpc *Server) SpiteStream(stream listenerrpc.ListenerRPC_SpiteStreamServer) error {
	listenerID, err := rpc.getListenerID(stream.Context())
	if err != nil {
		logs.Log.Error(err.Error())
		return err
	}
	listenersCh[listenerID] = stream
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		sess := core.Sessions.Get(msg.SessionId)
		if ch, ok := sess.GetTask(msg.TaskId); ok {
			ch <- msg.Spite
		}
	}
}
