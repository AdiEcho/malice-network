package rpc

import (
	"context"
	"errors"
	"github.com/chainreactors/malice-network/proto/client/commonpb"
	"github.com/chainreactors/malice-network/proto/services/clientrpc"
	"runtime"

	"github.com/chainreactors/malice-network/server/core"
	"github.com/chainreactors/malice-network/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"time"
)

var (
	// ErrInvalidBeaconID - Invalid Beacon ID in request
	ErrInvalidBeaconID = status.Error(codes.InvalidArgument, "Invalid beacon ID")
	// ErrInvalidBeaconTaskID - Invalid Beacon ID in request
	ErrInvalidBeaconTaskID = status.Error(codes.InvalidArgument, "Invalid beacon task ID")

	// ErrInvalidSessionID - Invalid Session ID in request
	ErrInvalidSessionID = status.Error(codes.InvalidArgument, "Invalid session ID")

	// ErrMissingRequestField - Returned when a request does not contain a commonpb.Request
	ErrMissingRequestField = status.Error(codes.InvalidArgument, "Missing session request field")
	// ErrAsyncNotSupported - Unsupported mode / command type
	ErrAsyncNotSupported = status.Error(codes.Unavailable, "Async not supported for this command")
	// ErrDatabaseFailure - Generic database failure error (real error is logged)
	ErrDatabaseFailure = status.Error(codes.Internal, "Database operation failed")

	// ErrInvalidName - Invalid name
	ErrInvalidName = status.Error(codes.InvalidArgument, "Invalid session name, alphanumerics and _-. only")

	//ErrInvalidBeaconTaskCancelState = status.Error(codes.InvalidArgument, fmt.Sprintf("Invalid task state, must be '%s' to cancel", models.PENDING))
)

type Server struct {
	// Magical methods to break backwards compatibility
	// Here be dragons: https://github.com/grpc/grpc-go/issues/3794
	clientrpc.UnimplementedMaliceRPCServer
}

// GenericRequest - Generic request interface to use with generic handlers
type GenericRequest interface {
	Reset()
	String() string
	ProtoMessage()
	ProtoReflect() protoreflect.Message

	GetRequest() *commonpb.Request
}

// GenericResponse - Generic response interface to use with generic handlers
type GenericResponse interface {
	Reset()
	String() string
	ProtoMessage()
	ProtoReflect() protoreflect.Message

	GetResponse() *commonpb.Response
}

// NewServer - Create new server instance
func NewServer() *Server {
	// todo event
	return &Server{}
}

// GenericHandler - Pass the request to the Sliver/Session
func (rpc *Server) GenericHandler(req GenericRequest, resp GenericResponse) error {
	var err error
	request := req.GetRequest()
	if request == nil {
		return ErrMissingRequestField
	}
	if request.Async {
		err = rpc.asyncGenericHandler(req, resp)
		return err
	}

	// Sync request
	session := core.Sessions.Get(request.SessionID)
	if session == nil {
		return ErrInvalidSessionID
	}

	// Overwrite unused implant fields before re-serializing
	request.SessionID = ""
	request.BeaconID = ""

	//reqData, err := proto.Marshal(req)
	//if err != nil {
	//	return err
	//}
	//
	//data, err := session.Request(sliverpb.MsgNumber(req), rpc.getTimeout(req), reqData)
	//if err != nil {
	//	return err
	//}
	//err = proto.Unmarshal(data, resp)
	//if err != nil {
	//	return err
	//}
	return rpc.getError(resp)
}

// asyncGenericHandler - Generic handler for async request/response's for beacon tasks
func (rpc *Server) asyncGenericHandler(req GenericRequest, resp GenericResponse) error {
	// VERY VERBOSE
	// rpcLog.Debugf("Async Generic Handler: %#v", req)
	//request := req.GetRequest()
	//if request == nil {
	//	return ErrMissingRequestField
	//}
	//
	//beacon, err := db.BeaconByID(request.BeaconID)
	//if beacon == nil || err != nil {
	//	rpcLog.Errorf("Invalid beacon ID in request: %s", err)
	//	return ErrInvalidBeaconID
	//}
	//
	//// Overwrite unused implant fields before re-serializing
	//request.SessionID = ""
	//request.BeaconID = ""
	//reqData, err := proto.Marshal(req)
	//if err != nil {
	//	return err
	//}
	//taskResponse := resp.GetResponse()
	//taskResponse.Async = true
	//taskResponse.BeaconID = beacon.ID.String()
	//task, err := beacon.Task(&sliverpb.Envelope{
	//	Type: sliverpb.MsgNumber(req),
	//	Data: reqData,
	//})
	//if err != nil {
	//	rpcLog.Errorf("Database error: %s", err)
	//	return ErrDatabaseFailure
	//}
	//parts := strings.Split(string(req.ProtoReflect().Descriptor().FullName().Name()), ".")
	//name := parts[len(parts)-1]
	//task.Description = name
	//err = db.Session().Save(task).Error
	//if err != nil {
	//	rpcLog.Errorf("Database error: %s", err)
	//	return ErrDatabaseFailure
	//}
	//taskResponse.TaskID = task.ID.String()
	//rpcLog.Debugf("Successfully tasked beacon: %#v", taskResponse)
	return nil
}

func (rpc *Server) GetBasicInfo(ctx context.Context, _ *commonpb.Empty) (*commonpb.Basic, error) {
	return &commonpb.Basic{
		Major: 0,
		Minor: 0,
		Patch: 1,
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
	}, nil
}

// getTimeout - Get the specified timeout from the request or the default
func (rpc *Server) getTimeout(req GenericRequest) time.Duration {
	timeout := req.GetRequest().Timeout
	if time.Duration(timeout) < time.Second {
		return utils.MinTimeout
	}
	return time.Duration(timeout)
}

// getError - Check an implant's response for Err and convert it to an `error` type
func (rpc *Server) getError(resp GenericResponse) error {
	respHeader := resp.GetResponse()
	if respHeader != nil && respHeader.Err != "" {
		return errors.New(respHeader.Err)
	}
	return nil
}
