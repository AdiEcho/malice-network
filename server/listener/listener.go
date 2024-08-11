package listener

import (
	"context"
	"fmt"
	"github.com/chainreactors/logs"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/helper/mtls"
	"github.com/chainreactors/malice-network/helper/types"
	"github.com/chainreactors/malice-network/proto/client/clientpb"
	"github.com/chainreactors/malice-network/proto/listener/lispb"
	"github.com/chainreactors/malice-network/proto/services/listenerrpc"
	"github.com/chainreactors/malice-network/server/internal/certs"
	"github.com/chainreactors/malice-network/server/internal/configs"
	"github.com/chainreactors/malice-network/server/internal/core"
	"github.com/chainreactors/malice-network/server/web"
	"google.golang.org/grpc"
	"net"
	"os"
	"strconv"
)

var (
	Listener *listener
)

func GenerateClientConfig(server *configs.ServerConfig, cfg *configs.ListenerConfig) (*mtls.ClientConfig, error) {
	var err error
	if _, err = os.Stat(cfg.Name + ".yaml"); os.IsExist(err) {
		config, err := mtls.ReadConfig(cfg.Name + ".yaml")
		if err != nil {
			return nil, err
		}
		return config, nil
	}
	clientConfig, err := certs.ClientGenerateCertificate(server.GRPCHost,
		cfg.Name, int(server.GRPCPort), certs.ListenerCA)
	if err != nil {
		return nil, err
	}

	err = mtls.WriteConfig(clientConfig, certs.ListenerNamespace, cfg.Name)
	if err != nil {
		return nil, err
	}
	return clientConfig, nil
}

func NewListener(clientConf *mtls.ClientConfig, cfg *configs.ListenerConfig) error {
	options, err := mtls.GetGrpcOptions([]byte(clientConf.CACertificate), []byte(clientConf.Certificate), []byte(clientConf.PrivateKey), certs.ListenerNamespace)
	if err != nil {
		return err
	}
	listenerCfg, err := mtls.ReadConfig(cfg.Auth)
	if err != nil {
		return err
	}
	serverAddress := net.JoinHostPort(listenerCfg.LHost, strconv.Itoa(listenerCfg.LPort))
	conn, err := grpc.Dial(serverAddress, options...)
	if err != nil {
		return err
	}

	lis := &listener{
		Rpc:       listenerrpc.NewListenerRPCClient(conn),
		Name:      cfg.Name,
		Host:      serverAddress,
		pipelines: make(core.Pipelines),
		conn:      conn,
		cfg:       cfg,
	}

	l := &lispb.Pipelines{}
	for _, tcpPipeline := range cfg.TcpPipelines {
		pipeline := &lispb.Pipeline{
			Body: &lispb.Pipeline_Tcp{
				Tcp: &lispb.TCPPipeline{
					Name: tcpPipeline.Name,
					Host: tcpPipeline.Host,
					Port: uint32(tcpPipeline.Port),
				},
			},
		}
		l.Pipelines = append(l.Pipelines, pipeline)
	}
	_, err = lis.Rpc.RegisterListener(context.Background(), &lispb.RegisterListener{
		Id:        fmt.Sprintf("%s_%s", lis.Name, lis.Host),
		Name:      lis.Name,
		Host:      lis.Host,
		Addr:      serverAddress,
		Pipelines: l,
	})
	if err != nil {
		return err
	}
	lis.Start()
	Listener = lis
	return nil
}

type listener struct {
	Rpc       listenerrpc.ListenerRPCClient
	Name      string
	Host      string
	pipelines core.Pipelines
	conn      *grpc.ClientConn
	cfg       *configs.ListenerConfig
}

func (lns *listener) ID() string {
	return fmt.Sprintf("%s_%s", lns.Name, lns.Host)
}

func (lns *listener) Start() {
	go lns.Handler()
	for _, tcp := range lns.cfg.TcpPipelines {
		pipeline, err := StartTcpPipeline(lns.conn, tcp)
		if err != nil {
			logs.Log.Errorf("Failed to start tcp pipeline %s", err)
			continue
		}
		logs.Log.Importantf("Started tcp pipeline %s, encryption: %t, tls: %t", pipeline.ID(), pipeline.Encryption.Enable, pipeline.TlsConfig.Enable)
		lns.registerPipeline(pipeline)
	}
	for _, website := range lns.cfg.Websites {
		if !website.Enable {
			continue
		}
		httpServer := web.NewHTTPServer(int(website.Port), website.RootPath, website.WebsiteName)
		go httpServer.Start()
	}
}

func (lns *listener) ToProtobuf() *clientpb.Listener {
	return &clientpb.Listener{
		Id:        lns.ID(),
		Pipelines: lns.pipelines.ToProtobuf(),
	}
}

func (lns *listener) Handler() {
	stream, err := lns.Rpc.JobStream(context.Background())
	if err != nil {
		return
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			return
		}
		var resp *clientpb.JobStatus
		switch msg.Ctrl {
		case consts.CtrlPipelineStart:
			resp = lns.startHandler(msg.Job)
		case consts.CtrlPipelineStop:
			resp = lns.stopHandler(msg.Job)
		}
		err = stream.Send(resp)
		if err != nil {
			return
		}
	}
}

func (lns *listener) registerPipeline(pipeline core.Pipeline) {
	lns.pipelines.Add(pipeline)
	lns.Rpc.RegisterPipeline(context.Background(), types.BuildPipeline(pipeline.ToProtobuf()))
}

func (lns *listener) startHandler(job *clientpb.Job) *clientpb.JobStatus {
	var err error
	pipeline := job.GetPipeline()
	switch pipeline.Body.(type) {
	case *lispb.Pipeline_Tcp:
		p := lns.pipelines.Get(pipeline.GetTcp().Name)
		if p == nil {
			tcpPipeline, err := StartTcpPipeline(lns.conn, ToTcpConfig(pipeline.GetTcp()))
			if err != nil {
				break
			}
			lns.registerPipeline(tcpPipeline)
		} else {
			err = p.Start()
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return &clientpb.JobStatus{
			ListenerId: lns.ID(),
			Ctrl:       consts.CtrlPipelineStart,
			Status:     consts.CtrlStatusFailed,
			Error:      err.Error(),
		}
	}
	return &clientpb.JobStatus{
		ListenerId: lns.ID(),
		Ctrl:       consts.CtrlPipelineStart,
		Status:     consts.CtrlStatusSuccess,
	}
}

func (lns *listener) stopHandler(job *clientpb.Job) *clientpb.JobStatus {
	var err error
	pipeline := job.GetPipeline()
	switch pipeline.Body.(type) {
	case *lispb.Pipeline_Tcp:
		p := lns.pipelines.Get(pipeline.GetTcp().Name)
		err = p.Close()
		if err != nil {
			break
		}

	}
	if err != nil {
		return &clientpb.JobStatus{
			ListenerId: lns.ID(),
			Ctrl:       consts.CtrlPipelineStop,
			Status:     consts.CtrlStatusFailed,
			Error:      err.Error(),
		}
	}
	return &clientpb.JobStatus{
		ListenerId: lns.ID(),
		Ctrl:       consts.CtrlPipelineStop,
		Status:     consts.CtrlStatusSuccess,
	}
}
