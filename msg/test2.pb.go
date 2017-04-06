// Code generated by protoc-gen-go.
// source: test2.proto
// DO NOT EDIT!

package msg

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// sample mseeage
type PingReq struct {
	Dummy            *uint32 `protobuf:"varint,1,req,name=dummy" json:"dummy,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PingReq) Reset()                    { *m = PingReq{} }
func (m *PingReq) String() string            { return proto.CompactTextString(m) }
func (*PingReq) ProtoMessage()               {}
func (*PingReq) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *PingReq) GetDummy() uint32 {
	if m != nil && m.Dummy != nil {
		return *m.Dummy
	}
	return 0
}

type PongAns struct {
	Err              *ErrorCode `protobuf:"varint,1,req,name=err,enum=msg.ErrorCode" json:"err,omitempty"`
	XXX_unrecognized []byte     `json:"-"`
}

func (m *PongAns) Reset()                    { *m = PongAns{} }
func (m *PongAns) String() string            { return proto.CompactTextString(m) }
func (*PongAns) ProtoMessage()               {}
func (*PongAns) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *PongAns) GetErr() ErrorCode {
	if m != nil && m.Err != nil {
		return *m.Err
	}
	return ErrorCode_ERR_SUCCESS
}

func init() {
	proto.RegisterType((*PingReq)(nil), "msg.PingReq")
	proto.RegisterType((*PongAns)(nil), "msg.PongAns")
}

func init() { proto.RegisterFile("test2.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 116 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2e, 0x49, 0x2d, 0x2e,
	0x31, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xce, 0x2d, 0x4e, 0x97, 0xe2, 0x02, 0x89,
	0x40, 0x04, 0x94, 0xe4, 0xb9, 0xd8, 0x03, 0x32, 0xf3, 0xd2, 0x83, 0x52, 0x0b, 0x85, 0x44, 0xb8,
	0x58, 0x53, 0x4a, 0x73, 0x73, 0x2b, 0x25, 0x18, 0x15, 0x98, 0x34, 0x78, 0x83, 0x20, 0x1c, 0x25,
	0x6d, 0x2e, 0xf6, 0x80, 0xfc, 0xbc, 0x74, 0xc7, 0xbc, 0x62, 0x21, 0x05, 0x2e, 0xe6, 0xd4, 0xa2,
	0x22, 0xb0, 0x34, 0x9f, 0x11, 0x9f, 0x5e, 0x6e, 0x71, 0xba, 0x9e, 0x6b, 0x51, 0x51, 0x7e, 0x91,
	0x73, 0x7e, 0x4a, 0x6a, 0x10, 0x48, 0x0a, 0x10, 0x00, 0x00, 0xff, 0xff, 0x52, 0xe0, 0x9d, 0xbe,
	0x6c, 0x00, 0x00, 0x00,
}
