package rpc

import (
	"context"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/malice-network/helper/mtls"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/commonpb"
	"github.com/chainreactors/malice-network/proto/listener/lispb"
	"github.com/chainreactors/malice-network/proto/services/listenerrpc"
	"github.com/chainreactors/malice-network/server/internal/certs"
	"github.com/chainreactors/malice-network/server/internal/core"
	"github.com/chainreactors/malice-network/server/internal/db"
	"github.com/chainreactors/malice-network/server/internal/db/models"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
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
	listenerID, err := getListenerID(stream.Context())
	if err != nil {
		logs.Log.Error(err.Error())
		return err
	}
	listenersCh[listenerID] = stream
	dbSession := db.Session()
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		sess, ok := core.Sessions.Get(msg.SessionId)

		err = models.UpdateLast(dbSession, sess.ID)
		if err != nil {
			logs.Log.Error(err.Error())
		}
		if !ok {
			return ErrNotFoundSession
		}
		if size := proto.Size(msg.Spite); size <= 1000 {
			logs.Log.Debugf("[server.%s] receive spite %s from %s, %v", sess.ID, msg.Spite.Name, msg.ListenerId, msg.Spite)
		} else {
			logs.Log.Debugf("[server.%s] receive spite %s from %s, %d bytes", sess.ID, msg.Spite.Name, msg.ListenerId, size)
		}

		if ch, ok := sess.GetResp(msg.TaskId); ok {
			ch <- msg.Spite
		}
	}
}

func (s *Server) AddListener(ctx context.Context, req *lispb.RegisterListener) (*commonpb.Empty, error) {
	_, _, err := certs.ClientGenerateCertificate(req.Host, req.Name, 5004, certs.ListenerCA)
	if err != nil {
		return &commonpb.Empty{}, err
	}
	return &commonpb.Empty{}, nil
}

func (s *Server) RemoveListener(ctx context.Context, req *lispb.RegisterListener) (*commonpb.Empty, error) {
	err := mtls.RemoveConfig(req.Name, certs.ListenerCA)
	if err != nil {
		return &commonpb.Empty{}, err
	}
	return &commonpb.Empty{}, nil
}

func (s *Server) ListListeners(ctx context.Context, req *commonpb.Empty) (*clientpb.Listeners, error) {
	files, err := mtls.GetListeners()
	if err != nil {
		return nil, err
	}
	listeners := &clientpb.Listeners{}
	for _, file := range files {
		listeners.Listeners = append(listeners.Listeners, &clientpb.Listener{
			Id: file,
		})
	}

	return listeners, nil
}
