package rpc

import (
	"context"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
)

var (
	RequeueDelay = time.Second
)

type appServer interface {
	Serve(ctx context.Context, typ string, req []byte) ([]byte, error)
}

type Server struct {
	ctx      context.Context
	srv      appServer
	producer *nsq.Producer
}

func NewServer(ctx context.Context, srv appServer, producer *nsq.Producer) *Server {
	return &Server{
		ctx:      ctx,
		srv:      srv,
		producer: producer,
	}
}

func (s *Server) HandleMessage(m *nsq.Message) error {
	// raspakiraj poruku u envelope
	eReq, err := NewEnvelope(m.Body)
	if err != nil {
		// TODO signalizirati da se nesto raspalo
		// return errors.Wrap(err, "envelope unpack failed")
		return nil
	}
	// provjeri da li je expired
	if eReq.Expired() {
		//log.S("type", eReq.Type).S("correlationId", eReq.CorrelationId).Info("expired")
		// TODO signalizirati da sam dobio expired poruku
		return nil
	}
	// radi request
	rsp, appErr := s.srv.Serve(s.ctx, eReq.Type, eReq.Body)
	// ako je timeout ili cancel
	if s.ctx.Err() != nil {
		m.RequeueWithoutBackoff(RequeueDelay)
		return nil
	}
	// treba li odgovoriti
	if eReq.ReplyTo == "" {
		return nil
	}
	// zapakuj
	eRsp := eReq.Reply(rsp, appErr)
	// posalji odgovor
	if err := s.producer.Publish(eReq.ReplyTo, eRsp.Bytes()); err != nil {
		return errors.Wrap(err, "nsq publish failed")
	}
	return nil
}

// TODO clean exit
// mozda i ne treba ako zatvori ctx ugasit ce sve sto je in process
// prije toga treba prestati primati poruke, ugasiti consumera
