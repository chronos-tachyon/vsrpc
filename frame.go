package vsrpc

import (
	"context"

	"github.com/chronos-tachyon/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ID uint32

type Method string

func MessageType(msg proto.Message) protoreflect.FullName {
	assert.NotNil(&msg)
	return msg.ProtoReflect().Descriptor().FullName()
}

type PacketReader interface {
	ReadPacket(ctx context.Context) ([]byte, func(), error)
}

func ReadFrame(ctx context.Context, r PacketReader, frame *Frame) error {
	assert.NotNil(&ctx)
	assert.NotNil(&r)
	assert.NotNil(&frame)
	frame.Reset()

	raw, dispose, err := r.ReadPacket(ctx)
	if err != nil {
		return err
	}
	defer dispose()

	err = proto.Unmarshal(raw, frame)
	if err != nil {
		frameType := MessageType(frame)
		return UnmarshalError{Type: frameType, Err: err}
	}
	return nil
}

type PacketWriter interface {
	WritePacket(ctx context.Context, p []byte) error
}

func WriteFrame(ctx context.Context, w PacketWriter, frame *Frame) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)
	assert.NotNil(&frame)

	raw, err := proto.Marshal(frame)
	if err != nil {
		return err
	}

	err = w.WritePacket(ctx, raw)
	if err != nil {
		return err
	}

	return nil
}

func WriteNoOp(ctx context.Context, w PacketWriter) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_NO_OP
	return WriteFrame(ctx, w, &frame)
}

func WriteBegin(ctx context.Context, w PacketWriter, id ID, method Method) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_BEGIN
	frame.CallId = uint32(id)
	frame.Method = string(method)
	if t, ok := ctx.Deadline(); ok {
		frame.Deadline = timestamppb.New(t)
	}
	return WriteFrame(ctx, w, &frame)
}

func WriteRequest(ctx context.Context, w PacketWriter, id ID, payload *anypb.Any) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_REQUEST
	frame.CallId = uint32(id)
	frame.Payload = payload
	return WriteFrame(ctx, w, &frame)
}

func WriteResponse(ctx context.Context, w PacketWriter, id ID, payload *anypb.Any) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_RESPONSE
	frame.CallId = uint32(id)
	frame.Payload = payload
	return WriteFrame(ctx, w, &frame)
}

func WriteHalfClose(ctx context.Context, w PacketWriter, id ID) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_HALF_CLOSE
	frame.CallId = uint32(id)
	return WriteFrame(ctx, w, &frame)
}

func WriteCancel(ctx context.Context, w PacketWriter, id ID) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_CANCEL
	frame.CallId = uint32(id)
	return WriteFrame(ctx, w, &frame)
}

func WriteEnd(ctx context.Context, w PacketWriter, id ID, status *Status) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var zero Status
	if status == nil {
		status = &zero
	}

	var frame Frame
	frame.Type = Frame_END
	frame.CallId = uint32(id)
	frame.Status = status
	return WriteFrame(ctx, w, &frame)
}

func WriteShutdown(ctx context.Context, w PacketWriter) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_SHUTDOWN
	return WriteFrame(ctx, w, &frame)
}

func WriteGoAway(ctx context.Context, w PacketWriter) error {
	assert.NotNil(&ctx)
	assert.NotNil(&w)

	var frame Frame
	frame.Type = Frame_GO_AWAY
	return WriteFrame(ctx, w, &frame)
}

func expectZeroCallId(frameType Frame_Type) bool {
	switch frameType {
	case Frame_NO_OP:
		fallthrough
	case Frame_SHUTDOWN:
		fallthrough
	case Frame_GO_AWAY:
		return true

	default:
		return false
	}
}

func expectNonZeroCallId(frameType Frame_Type) bool {
	switch frameType {
	case Frame_BEGIN:
		fallthrough
	case Frame_REQUEST:
		fallthrough
	case Frame_RESPONSE:
		fallthrough
	case Frame_HALF_CLOSE:
		fallthrough
	case Frame_CANCEL:
		fallthrough
	case Frame_END:
		return true

	default:
		return false
	}
}
