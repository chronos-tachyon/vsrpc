package vsrpc

import (
	"github.com/chronos-tachyon/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func MarshalAny(out *anypb.Any, in proto.Message) (*anypb.Any, error) {
	assert.NotNil(&out)

	if in == nil {
		return nil, nil
	}

	err := out.MarshalFrom(in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func UnmarshalAny(out proto.Message, in *anypb.Any) error {
	assert.NotNil(&out)
	proto.Reset(out)

	if in == nil {
		return nil
	}

	actual := in.MessageName()
	expect := MessageType(out)
	if actual != expect {
		return MessageTypeError{Actual: actual, Expect: expect}
	}

	if err := proto.Unmarshal(in.Value, out); err != nil {
		return UnmarshalError{Type: actual, Err: err}
	}
	return nil
}
