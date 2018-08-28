// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dapmsg.proto

package networkpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Dapmsg struct {
	Cmd                  string   `protobuf:"bytes,1,opt,name=cmd,proto3" json:"cmd,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	UnixTimeRecvd        int64    `protobuf:"varint,3,opt,name=unixTimeRecvd,proto3" json:"unixTimeRecvd,omitempty"`
	From                 string   `protobuf:"bytes,4,opt,name=from,proto3" json:"from,omitempty"`
	UniOrBroadcast       int64    `protobuf:"varint,5,opt,name=uniOrBroadcast,proto3" json:"uniOrBroadcast,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dapmsg) Reset()         { *m = Dapmsg{} }
func (m *Dapmsg) String() string { return proto.CompactTextString(m) }
func (*Dapmsg) ProtoMessage()    {}
func (*Dapmsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_a1733717f44737e9, []int{0}
}
func (m *Dapmsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Dapmsg.Unmarshal(m, b)
}
func (m *Dapmsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Dapmsg.Marshal(b, m, deterministic)
}
func (dst *Dapmsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Dapmsg.Merge(dst, src)
}
func (m *Dapmsg) XXX_Size() int {
	return xxx_messageInfo_Dapmsg.Size(m)
}
func (m *Dapmsg) XXX_DiscardUnknown() {
	xxx_messageInfo_Dapmsg.DiscardUnknown(m)
}

var xxx_messageInfo_Dapmsg proto.InternalMessageInfo

func (m *Dapmsg) GetCmd() string {
	if m != nil {
		return m.Cmd
	}
	return ""
}

func (m *Dapmsg) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Dapmsg) GetUnixTimeRecvd() int64 {
	if m != nil {
		return m.UnixTimeRecvd
	}
	return 0
}

func (m *Dapmsg) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}

func (m *Dapmsg) GetUniOrBroadcast() int64 {
	if m != nil {
		return m.UniOrBroadcast
	}
	return 0
}

func init() {
	proto.RegisterType((*Dapmsg)(nil), "networkpb.Dapmsg")
}

func init() { proto.RegisterFile("dapmsg.proto", fileDescriptor_dapmsg_a1733717f44737e9) }

var fileDescriptor_dapmsg_a1733717f44737e9 = []byte{
	// 157 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0x49, 0x2c, 0xc8,
	0x2d, 0x4e, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xcc, 0x4b, 0x2d, 0x29, 0xcf, 0x2f,
	0xca, 0x2e, 0x48, 0x52, 0x9a, 0xc0, 0xc8, 0xc5, 0xe6, 0x02, 0x96, 0x13, 0x12, 0xe0, 0x62, 0x4e,
	0xce, 0x4d, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0x31, 0x85, 0x84, 0xb8, 0x58, 0x52,
	0x12, 0x4b, 0x12, 0x25, 0x98, 0x14, 0x18, 0x35, 0x78, 0x82, 0xc0, 0x6c, 0x21, 0x15, 0x2e, 0xde,
	0xd2, 0xbc, 0xcc, 0x8a, 0x90, 0xcc, 0xdc, 0xd4, 0xa0, 0xd4, 0xe4, 0xb2, 0x14, 0x09, 0x66, 0x05,
	0x46, 0x0d, 0xe6, 0x20, 0x54, 0x41, 0x90, 0xce, 0xb4, 0xa2, 0xfc, 0x5c, 0x09, 0x16, 0xb0, 0x61,
	0x60, 0xb6, 0x90, 0x1a, 0x17, 0x5f, 0x69, 0x5e, 0xa6, 0x7f, 0x91, 0x53, 0x51, 0x7e, 0x62, 0x4a,
	0x72, 0x62, 0x71, 0x89, 0x04, 0x2b, 0x58, 0x2b, 0x9a, 0x68, 0x12, 0x1b, 0xd8, 0x91, 0xc6, 0x80,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x91, 0x50, 0xe1, 0x55, 0xb4, 0x00, 0x00, 0x00,
}
