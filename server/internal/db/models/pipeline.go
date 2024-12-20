package models

import (
	"fmt"
	"github.com/chainreactors/malice-network/helper/consts"
	"github.com/chainreactors/malice-network/helper/proto/client/clientpb"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	"time"
)

type TlsConfig struct {
	Enable bool   `gorm:"column:enable"`
	Cert   string `gorm:"column:cert"`
	Key    string `gorm:"column:key"`
}

type EncryptionConfig struct {
	Enable bool   `gorm:"column:enable"`
	Type   string `gorm:"column:type"`
	Key    string `gorm:"column:key"`
}

// Pipeline
type Pipeline struct {
	ID         uuid.UUID        `gorm:"primaryKey;->;<-:create;type:uuid;"`
	CreatedAt  time.Time        `gorm:"->;<-:create;"`
	ListenerID string           `gorm:"type:string;"`
	Name       string           `gorm:"unique,type:string"`
	WebPath    string           `gorm:"type:string;default:''"`
	IP         string           `gorm:"type:string;default:''"`
	Host       string           `config:"host"`
	Port       uint16           `config:"port"`
	Type       string           `gorm:"type:string;"`
	Parser     string           `gorm:"type:string;"`
	Enable     bool             `gorm:"type:boolean;"`
	Tls        TlsConfig        `gorm:"embedded;embeddedPrefix:tls_"`
	Encryption EncryptionConfig `gorm:"embedded;embeddedPrefix:encryption_"`
}

func (p *Pipeline) Address() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

// BeforeCreate - GORM hook
func (l *Pipeline) BeforeCreate(tx *gorm.DB) (err error) {
	l.ID, err = uuid.NewV4()
	if err != nil {
		return err
	}
	l.CreatedAt = time.Now()
	return nil
}

func FromPipelinePb(pipeline *clientpb.Pipeline, ip string) *Pipeline {
	switch body := pipeline.Body.(type) {
	case *clientpb.Pipeline_Tcp:
		return &Pipeline{
			ListenerID: pipeline.ListenerId,
			Name:       pipeline.Name,
			Enable:     pipeline.Enable,
			Parser:     pipeline.Parser,
			Host:       body.Tcp.Host,
			IP:         ip,
			Port:       uint16(body.Tcp.Port),
			Type:       consts.TCPPipeline,
			Tls:        ToTlsDB(pipeline.Tls),
			Encryption: ToEncryptionDB(pipeline.Encryption),
		}
	case *clientpb.Pipeline_Bind:
		return &Pipeline{
			ListenerID: pipeline.ListenerId,
			Name:       pipeline.Name,
			Enable:     pipeline.Enable,
			Parser:     pipeline.Parser,
			Type:       consts.BindPipeline,
			Tls:        ToTlsDB(pipeline.Tls),
			Encryption: ToEncryptionDB(pipeline.Encryption),
		}
	case *clientpb.Pipeline_Web:
		return &Pipeline{
			ListenerID: pipeline.ListenerId,
			Name:       pipeline.Name,
			Enable:     pipeline.Enable,
			Parser:     pipeline.Parser,
			WebPath:    body.Web.Root,
			IP:         ip,
			Port:       uint16(body.Web.Port),
			Type:       consts.WebsitePipeline,
			Tls:        ToTlsDB(pipeline.Tls),
		}
	default:
		return nil
	}
}

func ToPipelinePB(pipeline Pipeline) *clientpb.Pipeline {
	switch pipeline.Type {
	case consts.TCPPipeline:
		return &clientpb.Pipeline{
			Name:       pipeline.Name,
			ListenerId: pipeline.ListenerID,
			Enable:     pipeline.Enable,
			Parser:     pipeline.Parser,
			Body: &clientpb.Pipeline_Tcp{
				Tcp: &clientpb.TCPPipeline{
					Host: pipeline.Host,
					Port: uint32(pipeline.Port),
				},
			},
			Tls:        ToTlsProtobuf(&pipeline.Tls),
			Encryption: ToEncryptionProtobuf(&pipeline.Encryption),
		}
	case consts.BindPipeline:
		return &clientpb.Pipeline{
			Name:       pipeline.Name,
			ListenerId: pipeline.ListenerID,
			Enable:     pipeline.Enable,
			Parser:     pipeline.Parser,
			Body: &clientpb.Pipeline_Bind{
				Bind: &clientpb.BindPipeline{},
			},
			Tls:        ToTlsProtobuf(&pipeline.Tls),
			Encryption: ToEncryptionProtobuf(&pipeline.Encryption),
		}
	case consts.WebsitePipeline:
		return &clientpb.Pipeline{
			Name:       pipeline.Name,
			ListenerId: pipeline.ListenerID,
			Enable:     pipeline.Enable,
			Parser:     pipeline.Parser,
			Body: &clientpb.Pipeline_Web{
				Web: &clientpb.Website{
					Root: pipeline.WebPath,
					Port: uint32(pipeline.Port),
				},
			},
			Tls:        ToTlsProtobuf(&pipeline.Tls),
			Encryption: ToEncryptionProtobuf(&pipeline.Encryption),
		}
	default:
		return nil
	}
}

func ToTlsDB(tls *clientpb.TLS) TlsConfig {
	return TlsConfig{
		Cert:   tls.Cert,
		Key:    tls.Key,
		Enable: tls.Enable,
	}
}

func ToEncryptionDB(encryption *clientpb.Encryption) EncryptionConfig {
	return EncryptionConfig{
		Enable: encryption.Enable,
		Type:   encryption.Type,
		Key:    encryption.Key,
	}
}

func ToTlsProtobuf(tls *TlsConfig) *clientpb.TLS {
	return &clientpb.TLS{
		Enable: tls.Enable,
		Cert:   tls.Cert,
		Key:    tls.Key,
	}
}

func ToEncryptionProtobuf(encryption *EncryptionConfig) *clientpb.Encryption {
	return &clientpb.Encryption{
		Enable: encryption.Enable,
		Type:   encryption.Type,
		Key:    encryption.Key,
	}
}
