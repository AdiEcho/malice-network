package core

import (
	"errors"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/implant/commonpb"
	"github.com/chainreactors/malice-network/proto/listener/lispb"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var (
	// Sessions - Manages implant connections
	Sessions = &sessions{
		active: &sync.Map{},
	}

	// ErrUnknownMessageType - Returned if the implant did not understand the message for
	//                         example when the command is not supported on the platform
	ErrUnknownMessageType = errors.New("unknown message type")

	// ErrImplantSendTimeout - The implant did not respond prior to timeout deadline
	ErrImplantSendTimeout = errors.New("implant timeout")
)

func NewSession(req *lispb.RegisterSession) *Session {
	return &Session{
		Name:       req.RegisterData.Name,
		ProxyURL:   req.RegisterData.Proxy,
		Modules:    req.RegisterData.Module,
		Extensions: req.RegisterData.Extension,
		Filepath:   req.RegisterData.Filepath,
		ID:         req.SessionId,
		ListenerId: req.ListenerId,
		RemoteAddr: req.RemoteAddr,
		Os:         req.RegisterData.Os,
		Process:    req.RegisterData.Process,
		Timer:      req.RegisterData.Timer,
		Tasks:      &Tasks{active: &sync.Map{}},
		responses:  &sync.Map{},
	}
}

// Session - Represents a connection to an implant
type Session struct {
	ListenerId string
	ID         string
	Name       string
	RemoteAddr string
	Os         *commonpb.Os
	Process    *commonpb.Process
	Timer      *commonpb.Timer
	Filepath   string
	ActiveC2   string
	ProxyURL   string
	Modules    []string
	Extensions []string
	ConfigID   string
	PeerID     int64
	Locale     string
	Tasks      *Tasks // task manager
	taskseq    uint32
	responses  *sync.Map
}

func (s *Session) ToProtobuf() *clientpb.Session {
	return &clientpb.Session{
		SessionId: s.ID,
		Name:      s.Name,
		Os:        s.Os,
		Process:   s.Process,
		Timer:     s.Timer,
	}
}

func (s *Session) nextTaskId() uint32 {
	s.taskseq++
	return s.taskseq
}

func (s *Session) NewTask(name string, total int) *Task {
	s.taskseq++
	task := &Task{
		Type:      name,
		Total:     total,
		Id:        s.taskseq,
		SessionId: s.ID,
		done:      make(chan bool),
		end:       make(chan struct{}),
	}
	s.Tasks.Add(task)
	return task
}

func (s *Session) UpdateLastCheckin() {
	s.Timer.LastCheckin = uint64(time.Now().Unix())
}

// Request
func (s *Session) Request(msg *lispb.SpiteSession, stream grpc.ServerStream, timeout time.Duration) error {
	var err error
	done := make(chan struct{})
	go func() {
		err = stream.SendMsg(msg)
		if err != nil {
			logs.Log.Debugf(err.Error())
			return
		}
		close(done)
	}()
	select {
	case <-done:
		if err != nil {
			return err
		}
		return nil
	case <-time.After(timeout):
		return ErrImplantSendTimeout
	}
}

func (s *Session) RequestAndWait(msg *lispb.SpiteSession, stream grpc.ServerStream, timeout time.Duration) (*commonpb.Spite, error) {
	ch := make(chan *commonpb.Spite)
	s.StoreResp(msg.TaskId, ch)
	err := s.Request(msg, stream, timeout)
	if err != nil {
		return nil, err
	}
	resp := <-ch
	// todo @hyc save to database
	return resp, nil
}

// RequestWithStream - 'async' means that the response is not returned immediately, but is returned through the channel 'ch
func (s *Session) RequestWithStream(msg *lispb.SpiteSession, stream grpc.ServerStream, timeout time.Duration) (chan *commonpb.Spite, chan *commonpb.Spite, error) {
	respCh := make(chan *commonpb.Spite)
	s.StoreResp(msg.TaskId, respCh)
	err := s.Request(msg, stream, timeout)
	if err != nil {
		return nil, nil, err
	}

	in := make(chan *commonpb.Spite)
	go func() {
		defer close(respCh)
		var c = 0
		for spite := range in {
			err := stream.SendMsg(&lispb.SpiteSession{
				SessionId: s.ID,
				TaskId:    msg.TaskId,
				Spite:     spite,
			})
			if err != nil {
				logs.Log.Debugf(err.Error())
				return
			}
			logs.Log.Debugf("send message %s, %d", spite.Name, c)
			c++
		}
	}()
	return in, respCh, nil
}

func (s *Session) RequestWithAsync(msg *lispb.SpiteSession, stream grpc.ServerStream, timeout time.Duration) (chan *commonpb.Spite, error) {
	respCh := make(chan *commonpb.Spite)
	s.StoreResp(msg.TaskId, respCh)
	err := s.Request(msg, stream, timeout)
	if err != nil {
		return nil, err
	}

	return respCh, nil
}

func (s *Session) StoreResp(taskId uint32, ch chan *commonpb.Spite) {
	s.responses.Store(taskId, ch)
}

func (s *Session) GetResp(taskId uint32) (chan *commonpb.Spite, bool) {
	msg, ok := s.responses.Load(taskId)
	if !ok {
		return nil, false
	}
	return msg.(chan *commonpb.Spite), true
}

func (s *Session) DeleteResp(taskId uint32) {
	ch, ok := s.GetResp(taskId)
	if ok {
		close(ch)
	}
	s.responses.Delete(taskId)
}

// sessions - Manages the slivers, provides atomic access
type sessions struct {
	active *sync.Map // map[uint32]*Session
}

// All - Return a list of all sessions
func (s *sessions) All() []*Session {
	all := []*Session{}
	s.active.Range(func(key, value interface{}) bool {
		all = append(all, value.(*Session))
		return true
	})
	return all
}

// Get - Get a session by ID
func (s *sessions) Get(sessionID string) (*Session, bool) {
	if val, ok := s.active.Load(sessionID); ok {
		return val.(*Session), true
	}
	return nil, false
}

// Add - Add a sliver to the hive (atomically)
func (s *sessions) Add(session *Session) *Session {
	s.active.Store(session.ID, session)
	//EventBroker.Publish(Event{
	//	EventType: consts.SessionOpenedEvent,
	//	Session:   session,
	//})
	return session
}

// Remove - Remove a sliver from the hive (atomically)
func (s *sessions) Remove(sessionID string) {
	val, ok := s.active.Load(sessionID)
	if !ok {
		return
	}
	parentSession := val.(*Session)
	//children := findAllChildrenByPeerID(parentSession.PeerID)
	s.active.Delete(parentSession.ID)
	//coreLog.Debugf("Removing %d children of session %d (%v)", len(children), parentSession.ID, children)
	//for _, child := range children {
	//	childSession, ok := s.active.LoadAndDelete(child.SessionID)
	//	if ok {
	//		PivotSessions.Delete(childSession.(*Session).Connection.ID)
	//		EventBroker.Publish(Event{
	//			EventType: consts.SessionClosedEvent,
	//			Session:   childSession.(*Session),
	//		})
	//	}
	//}

	//Remove the parent session
	//EventBroker.Publish(Event{
	//	EventType: consts.SessionClosedEvent,
	//	Session:   parentSession,
	//})
}
