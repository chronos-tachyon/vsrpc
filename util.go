package vsrpc

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type void struct{}

type deadlineSetter interface {
	SetDeadline(time.Time) error
}

func neverPanic(fn func()) {
	if fn == nil {
		return
	}

	defer recover()
	fn()
}

func try(fn func() error) (err error) {
	defer func() {
		panicValue := recover()
		switch x := panicValue.(type) {
		case nil:
			return

		case error:
			err = x

		case string:
			err = errors.New(x)

		default:
			err = fmt.Errorf("%v", panicValue)
		}
		err = PanicError{Err: err}
	}()
	err = fn()
	return
}

const emptyTypeURL = "type.googleapis.com/google.protobuf.Empty"

func asAny(in proto.Message) (*anypb.Any, error) {
	if in == nil {
		return &anypb.Any{TypeUrl: emptyTypeURL}, nil
	}

	if _, ok := in.(*emptypb.Empty); ok {
		return &anypb.Any{TypeUrl: emptyTypeURL}, nil
	}

	if x, ok := in.(*anypb.Any); ok {
		return x, nil
	}

	return anypb.New(in)
}
