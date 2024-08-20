package rpc

import (
	"context"
	"github.com/chainreactors/malice-network/helper/types"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"
	"github.com/chainreactors/malice-network/server/internal/core"
	"github.com/chainreactors/malice-network/server/internal/db"
)

func (rpc *Server) GetSessions(ctx context.Context, _ *clientpb.Empty) (*clientpb.Sessions, error) {
	sessions := &clientpb.Sessions{}
	for _, session := range core.Sessions.All() {
		sessions.Sessions = append(sessions.Sessions, session.ToProtobuf())
	}
	return sessions, nil
}

func (rpc *Server) GetAlivedSessions(ctx context.Context, _ *clientpb.Empty) (*clientpb.Sessions, error) {
	var sessions []*clientpb.Session
	for _, session := range core.Sessions.All() {
		sessionProto := session.ToProtobuf()
		if sessionProto.IsAlive {
			sessions = append(sessions, session.ToProtobuf())
		}
	}
	return &clientpb.Sessions{Sessions: sessions}, nil
}

func (rpc *Server) GetSession(ctx context.Context, req *clientpb.SessionRequest) (*clientpb.Session, error) {
	session, ok := core.Sessions.Get(req.SessionId)
	if ok {
		return nil, ErrNotFoundSession
	}
	return session.ToProtobuf(), nil
}

func (rpc *Server) BasicSessionOP(ctx context.Context, req *clientpb.BasicUpdateSession) (*clientpb.Empty, error) {
	if req.IsDelete {
		core.Sessions.Remove(req.SessionId)
		err := db.DeleteSession(req.SessionId)
		if err != nil {
			return nil, err
		}
	} else {
		session, ok := core.Sessions.Get(req.SessionId)
		if !ok {
			return nil, ErrNotFoundSession
		}
		if req.Note != "" {
			session.Name = req.Note
		}
		if req.GroupName != "" {
			session.Group = req.GroupName
		}
		err := db.UpdateSession(req.SessionId, req.Note, req.GroupName)
		if err != nil {
			return nil, err
		}
	}
	return &clientpb.Empty{}, nil
}

func (rpc *Server) Info(ctx context.Context, req *implantpb.Request) (*clientpb.Task, error) {
	greq, err := newGenericRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	ch, err := rpc.asyncGenericHandler(ctx, greq)
	if err != nil {
		return nil, err
	}

	go greq.HandlerAsyncResponse(ch, types.MsgSysInfo)
	return greq.Task.ToProtobuf(), nil
}
