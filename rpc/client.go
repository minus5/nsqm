package rpc

import (
	"context"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
)

type Client struct {
	publisher *nsq.Producer
	reqTopic  string
	rspTopic  string
	msgNo     int
	s         map[string]chan *Envelope
	sync.Mutex
}

func NewClient(publisher *nsq.Producer, reqTopic, rspTopic string) *Client {
	rand.Seed(time.Now().UnixNano())
	return &Client{
		publisher: publisher,
		reqTopic:  reqTopic,
		rspTopic:  rspTopic,
		msgNo:     rand.Intn(math.MaxInt32),
		s:         make(map[string]chan *Envelope),
	}
}

func (c *Client) HandleMessage(m *nsq.Message) error {
	e, err := NewEnvelope(m.Body)
	if err != nil {
		// TODO poruka ne valja
		return nil
	}
	if s, found := c.get(e.CorrelationId); found {
		// when s == nil, means that request timed out, nobody is waiting for response
		// nothing to do in that case
		if s != nil {
			s <- e
		}
		return nil
	}
	//log.S("id", e.CorrelationId).Info("subscriber not found")
	// TODO logiranje
	return nil
}

func (c *Client) Call(ctx context.Context, typ string, req []byte) ([]byte, string, error) {
	correlationId := c.correlationID()
	eReq := &Envelope{
		Type:          typ,
		ReplyTo:       c.rspTopic,
		CorrelationId: correlationId,
		Body:          req,
	}
	if d, ok := ctx.Deadline(); ok {
		eReq.ExpiresAt = d.Unix()
	}
	rspCh := make(chan *Envelope)
	c.add(correlationId, rspCh)

	if err := c.publisher.Publish(c.reqTopic, eReq.Bytes()); err != nil {
		return nil, "", errors.Wrap(err, "nsq publish failed")
	}

	select {
	case rsp := <-rspCh:
		return rsp.Body, rsp.Error, nil
	case <-ctx.Done():
		c.timeout(correlationId)
		return nil, "", ctx.Err()
	}

}

func (c *Client) Close() {
	//t.publisher.
}

func (c *Client) correlationID() string {
	c.Lock()
	defer c.Unlock()
	if c.msgNo == math.MaxInt32 {
		c.msgNo = math.MinInt32
	} else {
		c.msgNo++
	}
	return strconv.Itoa(c.msgNo)
}

func (c *Client) add(id string, ch chan *Envelope) {
	c.Lock()
	defer c.Unlock()
	c.s[id] = ch
}

func (c *Client) get(id string) (chan *Envelope, bool) {
	c.Lock()
	defer c.Unlock()
	ch, ok := c.s[id]
	if ok {
		delete(c.s, id)
	}
	return ch, ok
}

func (c *Client) timeout(id string) {
	c.Lock()
	defer c.Unlock()
	if _, found := c.s[id]; found {
		c.s[id] = nil
	}
}
