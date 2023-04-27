// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v4.22.3
// source: vsrpc/frame.proto

package vsrpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Frame_Type int32

const (
	// NO_OP denotes a frame that does nothing.
	//
	// A NO_OP frame is *not* a ping and does *not* solicit a response.  Both
	// clients and servers should ignore any NO_OP frames they receive, and
	// neither clients nor servers should send NO_OP frames under normal
	// circumstances.
	//
	// Direction: any
	// Required fields: type
	// Optional fields: NONE
	Frame_NO_OP Frame_Type = 0
	// SHUTDOWN tells the server that no more BEGIN frames will be sent.
	//
	// Direction: client to server
	// Required fields: type
	// Optional fields: NONE
	Frame_SHUTDOWN Frame_Type = 1
	// GO_AWAY tells the client that future BEGIN frames will not be honored,
	// because the server is shutting down the connection.
	//
	// The server may start aborting in-progress RPC calls if they do not
	// complete quickly enough, or it may close the connection without warning,
	// with or without sending a GO_AWAY frame.  The client must treat
	// abandoned RPC calls as having failed with status code ABORTED.
	//
	// Direction: server to client
	// Required fields: type
	// Optional fields: NONE
	Frame_GO_AWAY Frame_Type = 2
	// BEGIN creates a new RPC call.
	//
	// Direction: client to server
	// Required fields: type, call_id, method
	// Optional fields: deadline
	Frame_BEGIN Frame_Type = 3
	// REQUEST sends a request body for an RPC call.
	//
	// Direction: client to server
	// Required fields: type, call_id, payload
	// Optional fields: NONE
	Frame_REQUEST Frame_Type = 4
	// RESPONSE sends a response body for an RPC call.
	//
	// It is valid for the server to send RESPONSE frames before the client has
	// sent HALF_CLOSE; this is how bidirectional streaming works.
	//
	// Direction: server to client
	// Required fields: type, call_id, payload
	// Optional fields: NONE
	Frame_RESPONSE Frame_Type = 5
	// HALF_CLOSE tells the server that no more REQUEST frames will be sent for
	// this RPC call.
	//
	// Direction: client to server
	// Required fields: type, call_id
	// Optional fields: NONE
	Frame_HALF_CLOSE Frame_Type = 6
	// CANCEL asks the server to cancel the in-progress RPC call.
	//
	// The RPC call remains outstanding until the server sends a END frame.
	// Status code CANCELLED indicates that the cancellation was accepted.
	//
	// Direction: client to server
	// Required fields: type, call_id
	// Optional fields: NONE
	Frame_CANCEL Frame_Type = 7
	// END reports the final status of the RPC call.
	//
	// If the client continues to send frames relating to this call, the server
	// must ignore them; they are NOT a protocol violation.
	//
	// Direction: server to client
	// Required fields: type, call_id, status
	// Optional fields: NONE
	Frame_END Frame_Type = 8
)

// Enum value maps for Frame_Type.
var (
	Frame_Type_name = map[int32]string{
		0: "NO_OP",
		1: "SHUTDOWN",
		2: "GO_AWAY",
		3: "BEGIN",
		4: "REQUEST",
		5: "RESPONSE",
		6: "HALF_CLOSE",
		7: "CANCEL",
		8: "END",
	}
	Frame_Type_value = map[string]int32{
		"NO_OP":      0,
		"SHUTDOWN":   1,
		"GO_AWAY":    2,
		"BEGIN":      3,
		"REQUEST":    4,
		"RESPONSE":   5,
		"HALF_CLOSE": 6,
		"CANCEL":     7,
		"END":        8,
	}
)

func (x Frame_Type) Enum() *Frame_Type {
	p := new(Frame_Type)
	*p = x
	return p
}

func (x Frame_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Frame_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_vsrpc_frame_proto_enumTypes[0].Descriptor()
}

func (Frame_Type) Type() protoreflect.EnumType {
	return &file_vsrpc_frame_proto_enumTypes[0]
}

func (x Frame_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Frame_Type.Descriptor instead.
func (Frame_Type) EnumDescriptor() ([]byte, []int) {
	return file_vsrpc_frame_proto_rawDescGZIP(), []int{0, 0}
}

type Frame struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     Frame_Type             `protobuf:"varint,1,opt,name=type,proto3,enum=vsrpc.Frame_Type" json:"type,omitempty"`
	CallId   uint32                 `protobuf:"varint,2,opt,name=call_id,json=callId,proto3" json:"call_id,omitempty"`
	Method   string                 `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Deadline *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=deadline,proto3" json:"deadline,omitempty"`
	Payload  *anypb.Any             `protobuf:"bytes,5,opt,name=payload,proto3" json:"payload,omitempty"`
	Status   *Status                `protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *Frame) Reset() {
	*x = Frame{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vsrpc_frame_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Frame) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Frame) ProtoMessage() {}

func (x *Frame) ProtoReflect() protoreflect.Message {
	mi := &file_vsrpc_frame_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Frame.ProtoReflect.Descriptor instead.
func (*Frame) Descriptor() ([]byte, []int) {
	return file_vsrpc_frame_proto_rawDescGZIP(), []int{0}
}

func (x *Frame) GetType() Frame_Type {
	if x != nil {
		return x.Type
	}
	return Frame_NO_OP
}

func (x *Frame) GetCallId() uint32 {
	if x != nil {
		return x.CallId
	}
	return 0
}

func (x *Frame) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Frame) GetDeadline() *timestamppb.Timestamp {
	if x != nil {
		return x.Deadline
	}
	return nil
}

func (x *Frame) GetPayload() *anypb.Any {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Frame) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

var File_vsrpc_frame_proto protoreflect.FileDescriptor

var file_vsrpc_frame_proto_rawDesc = []byte{
	0x0a, 0x11, 0x76, 0x73, 0x72, 0x70, 0x63, 0x2f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x76, 0x73, 0x72, 0x70, 0x63, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x76, 0x73, 0x72, 0x70, 0x63, 0x2f, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe7, 0x02, 0x0a, 0x05, 0x46,
	0x72, 0x61, 0x6d, 0x65, 0x12, 0x25, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x11, 0x2e, 0x76, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65,
	0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x63,
	0x61, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x63, 0x61,
	0x6c, 0x6c, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x36, 0x0a, 0x08,
	0x64, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x64, 0x65, 0x61, 0x64,
	0x6c, 0x69, 0x6e, 0x65, 0x12, 0x2e, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x12, 0x25, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x76, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x77, 0x0a, 0x04, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x4e, 0x4f, 0x5f, 0x4f, 0x50, 0x10, 0x00, 0x12, 0x0c,
	0x0a, 0x08, 0x53, 0x48, 0x55, 0x54, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07,
	0x47, 0x4f, 0x5f, 0x41, 0x57, 0x41, 0x59, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x42, 0x45, 0x47,
	0x49, 0x4e, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10,
	0x04, 0x12, 0x0c, 0x0a, 0x08, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x10, 0x05, 0x12,
	0x0e, 0x0a, 0x0a, 0x48, 0x41, 0x4c, 0x46, 0x5f, 0x43, 0x4c, 0x4f, 0x53, 0x45, 0x10, 0x06, 0x12,
	0x0a, 0x0a, 0x06, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x10, 0x07, 0x12, 0x07, 0x0a, 0x03, 0x45,
	0x4e, 0x44, 0x10, 0x08, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x6f, 0x73, 0x2d, 0x74, 0x61, 0x63, 0x68, 0x79,
	0x6f, 0x6e, 0x2f, 0x76, 0x73, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_vsrpc_frame_proto_rawDescOnce sync.Once
	file_vsrpc_frame_proto_rawDescData = file_vsrpc_frame_proto_rawDesc
)

func file_vsrpc_frame_proto_rawDescGZIP() []byte {
	file_vsrpc_frame_proto_rawDescOnce.Do(func() {
		file_vsrpc_frame_proto_rawDescData = protoimpl.X.CompressGZIP(file_vsrpc_frame_proto_rawDescData)
	})
	return file_vsrpc_frame_proto_rawDescData
}

var file_vsrpc_frame_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_vsrpc_frame_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_vsrpc_frame_proto_goTypes = []interface{}{
	(Frame_Type)(0),               // 0: vsrpc.Frame.Type
	(*Frame)(nil),                 // 1: vsrpc.Frame
	(*timestamppb.Timestamp)(nil), // 2: google.protobuf.Timestamp
	(*anypb.Any)(nil),             // 3: google.protobuf.Any
	(*Status)(nil),                // 4: vsrpc.Status
}
var file_vsrpc_frame_proto_depIdxs = []int32{
	0, // 0: vsrpc.Frame.type:type_name -> vsrpc.Frame.Type
	2, // 1: vsrpc.Frame.deadline:type_name -> google.protobuf.Timestamp
	3, // 2: vsrpc.Frame.payload:type_name -> google.protobuf.Any
	4, // 3: vsrpc.Frame.status:type_name -> vsrpc.Status
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_vsrpc_frame_proto_init() }
func file_vsrpc_frame_proto_init() {
	if File_vsrpc_frame_proto != nil {
		return
	}
	file_vsrpc_status_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_vsrpc_frame_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Frame); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_vsrpc_frame_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_vsrpc_frame_proto_goTypes,
		DependencyIndexes: file_vsrpc_frame_proto_depIdxs,
		EnumInfos:         file_vsrpc_frame_proto_enumTypes,
		MessageInfos:      file_vsrpc_frame_proto_msgTypes,
	}.Build()
	File_vsrpc_frame_proto = out.File
	file_vsrpc_frame_proto_rawDesc = nil
	file_vsrpc_frame_proto_goTypes = nil
	file_vsrpc_frame_proto_depIdxs = nil
}