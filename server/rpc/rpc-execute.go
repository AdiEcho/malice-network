package rpc

import (
	"context"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/helper/types"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/implantpb"

	"github.com/chainreactors/malice-network/server/internal/core"
)

func (rpc *Server) Execute(ctx context.Context, req *implantpb.ExecRequest) (*clientpb.Task, error) {
	greq, err := newGenericRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	ch, err := rpc.asyncGenericHandler(ctx, greq)
	if err != nil {
		return nil, err
	}

	go func() {
		resp := <-ch

		err := AssertStatusAndResponse(resp, types.MsgExec)
		if err != nil {
			core.EventBroker.Publish(buildErrorEvent(greq.Task, err))
			return
		}
		greq.SetCallback(func() {
			greq.Task.Spite = resp
			core.EventBroker.Publish(core.Event{
				EventType: consts.EventTaskCallback,
				Task:      greq.Task,
			})
		})
		greq.Task.Done()
	}()
	return greq.Task.ToProtobuf(), nil
}

func (rpc *Server) ExecuteAssembly(ctx context.Context, req *implantpb.ExecuteAssembly) (*clientpb.Task, error) {
	greq, err := newGenericRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	ch, err := rpc.asyncGenericHandler(ctx, greq)
	if err != nil {
		return nil, err
	}
	go func() {
		resp := <-ch
		err := AssertStatusAndResponse(resp, types.MsgAssemblyResponse)
		if err != nil {
			core.EventBroker.Publish(buildErrorEvent(greq.Task, err))
			return
		}
		greq.SetCallback(func() {
			greq.Task.Spite = resp
			core.EventBroker.Publish(core.Event{
				EventType: consts.EventTaskCallback,
				Task:      greq.Task,
			})
		})
		greq.Task.Done()
	}()
	return greq.Task.ToProtobuf(), nil
}

func (rpc *Server) ExecuteShellcode(ctx context.Context, req *implantpb.ExecuteShellcode) (*clientpb.Task, error) {
	greq, err := newGenericRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	ch, err := rpc.asyncGenericHandler(ctx, greq)
	if err != nil {
		return nil, err
	}
	go func() {
		resp := <-ch
		err := AssertStatusAndResponse(resp, types.MsgAssemblyResponse)
		if err != nil {
			core.EventBroker.Publish(buildErrorEvent(greq.Task, err))
			return
		}
		greq.SetCallback(func() {
			greq.Task.Spite = resp
			core.EventBroker.Publish(core.Event{
				EventType: consts.EventTaskCallback,
				Task:      greq.Task,
			})
		})
		greq.Task.Done()
	}()
	return greq.Task.ToProtobuf(), nil
}

func (rpc *Server) ExecuteBof(ctx context.Context, req *implantpb.ExecuteBof) (*clientpb.Task, error) {
	greq, err := newGenericRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	ch, err := rpc.asyncGenericHandler(ctx, greq)
	if err != nil {
		return nil, err
	}
	go func() {
		resp := <-ch
		err := AssertStatusAndResponse(resp, types.MsgAssemblyResponse)
		if err != nil {
			core.EventBroker.Publish(buildErrorEvent(greq.Task, err))
			return
		}
		greq.SetCallback(func() {
			greq.Task.Spite = resp
			core.EventBroker.Publish(core.Event{
				EventType: consts.EventTaskCallback,
				Task:      greq.Task,
			})
		})
		greq.Task.Done()
	}()
	return greq.Task.ToProtobuf(), nil
}
