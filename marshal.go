package vsrpc

import (
	"github.com/chronos-tachyon/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func MarshalAny(out *anypb.Any, in proto.Message) error {
	assert.NotNil(&out)
	return out.MarshalFrom(in)
}

func UnmarshalAny(out proto.Message, in *anypb.Any) error {
	assert.NotNil(&out)
	return in.UnmarshalTo(out)
}
