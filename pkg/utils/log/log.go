package log

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func LogWithGRPC(ctx context.Context) *logrus.Entry {
	e := logrus.NewEntry(logrus.New())
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

func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}

func Info(args ...interface{}) {
	logrus.Info(args)
}
