package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeRequest(t *testing.T) {
	expiresAt := int64(1515350225)
	e := &Envelope{
		Method:        "Add",
		ReplyTo:       "service.rsp",
		CorrelationID: 12345,
		ExpiresAt:     expiresAt,
		Body:          []byte("iso medo u ducan"),
	}

	buf := e.encode(0)
	// fmt.Println(buf)
	// fmt.Println(string(buf))

	e2, err := decode(buf)
	assert.Nil(t, err)

	assert.Equal(t, e.Method, e2.Method)
	assert.Equal(t, e.ReplyTo, e2.ReplyTo)
	assert.Equal(t, e.CorrelationID, e2.CorrelationID)
	assert.Equal(t, e.ExpiresAt, e2.ExpiresAt)
	assert.Equal(t, e.Body, e2.Body)
	//assert.ObjectsAreEqual(e, e2)

	// b, _ := json.Marshal(e)
	// fmt.Printf("%s\n", b)
	// b, _ = json.Marshal(e2)
	// fmt.Printf("%s\n", b)
}

func TestEncodeReply(t *testing.T) {
	e := &Envelope{
		Error:         "overflow",
		CorrelationID: 12345,
		Body:          []byte("iso medo u ducan"),
	}

	buf := e.encode(1)
	// fmt.Println(buf)
	// fmt.Println(string(buf))

	e2, err := decode(buf)
	assert.Nil(t, err)
	assert.Equal(t, e.CorrelationID, e2.CorrelationID)
	assert.Equal(t, e.Error, e2.Error)
	assert.Equal(t, e.Body, e2.Body)

	// b, _ := json.Marshal(e)
	// fmt.Printf("%s\n", b)
	// b, _ = json.Marshal(e2)
	// fmt.Printf("%s\n", b)
}
