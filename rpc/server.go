package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
)

var (
	requeueDelay = time.Second
)

type appServer interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}

// Server rpc server side.
type Server struct {
	ctx      context.Context
	srv      appServer
	producer *nsq.Producer
}

// NewServer creates new rpc server for appServer.
// producer will be used for sending replies.
func NewServer(ctx context.Context, srv appServer, producer *nsq.Producer) *Server {
	return &Server{
		ctx:      ctx,
		srv:      srv,
		producer: producer,
	}
}

// HandleMessage server side handler.
func (s *Server) HandleMessage(m *nsq.Message) error {
	fin := func() {
		m.DisableAutoResponse()
		m.Finish()
	}
	// decode message
	req, err := Decode(m.Body)
	if err != nil {
		fin() // raise error without message requeue
		return errors.Wrap(err, "envelope unpack failed")
	}
	// check expiration
	if req.Expired() {
		fin()
		return fmt.Errorf("expired %s %d", req.Method, req.CorrelationID)
	}
	// call aplication
	appRsp, appErr := s.srv.Serve(s.ctx, req.Method, req.Body)
	// context timeout/cancel ?
	if s.ctx.Err() != nil {
		m.RequeueWithoutBackoff(requeueDelay)
		return nil
	}
	// need to reply
	if req.ReplyTo == "" {
		return nil
	}
	// create reply
	rsp := req.Reply(appRsp, appErr)
	// send reply
	if err := s.producer.Publish(req.ReplyTo, rsp.Encode()); err != nil {
		return errors.Wrap(err, "nsq publish failed")
	}
	return nil
}
