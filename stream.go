package vsrpc

import (
	"context"

	"github.com/chronos-tachyon/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BaseStream interface {
	Call() *Call
	Conn() *Conn
	Context() context.Context
	ID() ID
	Method() Method
	Cancel() error
}

type SendStream[T proto.Message] interface {
	BaseStream
	Send(in T) error
	CloseSend() error
}

type RecvStream[T proto.Message] interface {
	BaseStream
	Recv(blocking bool, out T) (ok bool, done bool, err error)
}

type BiStream[T proto.Message, U proto.Message] interface {
	BaseStream
	SendStream[T]
	RecvStream[U]
}

func NewStream[T proto.Message, U proto.Message](call *Call) BiStream[T, U] {
	return &implStream[T, U]{call: call}
}

func NewSendStream[T proto.Message](call *Call) SendStream[T] {
	return NewStream[T, *emptypb.Empty](call)
}

func NewRecvStream[T proto.Message](call *Call) RecvStream[T] {
	return NewStream[*emptypb.Empty, T](call)
}

func NewBiStream[T proto.Message, U proto.Message](call *Call) BiStream[T, U] {
	return NewStream[T, U](call)
}

type implStream[T proto.Message, U proto.Message] struct {
	call *Call
}

func (stream implStream[T, U]) Call() *Call {
	return stream.call
}

func (stream implStream[T, U]) Conn() *Conn {
	return stream.Call().Conn()
}

func (stream implStream[T, U]) Queue() *Queue {
	return stream.Call().Queue()
}

func (stream implStream[T, U]) Context() context.Context {
	return stream.Call().Context()
}

func (stream implStream[T, U]) ID() ID {
	return stream.Call().ID()
}

func (stream implStream[T, U]) Method() Method {
	return stream.Call().Method()
}

func (stream implStream[T, U]) Cancel() error {
	return stream.Call().Cancel()
}

func (stream implStream[T, U]) Close() error {
	return stream.Call().Close()
}

func (stream implStream[T, U]) Send(in T) error {
	assert.NotNil(&in)

	var tmp anypb.Any
	out, err := MarshalAny(&tmp, in)
	if err != nil {
		return err
	}
	return stream.Call().Send(out)
}

func (stream implStream[T, U]) CloseSend() error {
	return stream.Call().CloseSend()
}

func (stream implStream[T, U]) Recv(blocking bool, out U) (ok bool, done bool, err error) {
	assert.NotNil(&out)
	proto.Reset(out)

	var payload *anypb.Any
	payload, ok, done = stream.Queue().Recv(blocking)
	if ok && payload != nil {
		err = UnmarshalAny(out, payload)
		if err != nil {
			ok = false
		}
	}
	return
}
