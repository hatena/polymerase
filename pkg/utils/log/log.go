package log

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func LogWithGRPC(ctx context.Context) *log.Entry {
	e := log.NewEntry(log.New())
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			e = e.WithField(k, v)
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		if p.Addr != nil {
			e = e.WithField("peer_addr", p.Addr)
		}
		if p.AuthInfo != nil {
			e = e.WithField("peer_authinfo", p.AuthInfo)
		}
	}
	return e
}

func WithField(key string, value interface{}) *log.Entry {
	return log.WithField(key, value)
}

func Info(args ...interface{}) {
	log.Info(args)
}
