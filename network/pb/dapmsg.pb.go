// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/dappley/go-dappley/network/pb/dapmsg.proto

package networkpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import pb "github.com/dappley/go-dappley/core/pb"

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
	UnixTimeReceived     int64    `protobuf:"varint,3,opt,name=unix_time_received,json=unixTimeReceived,proto3" json:"unix_time_received,omitempty"`
	Key                  string   `protobuf:"bytes,4,opt,name=key,proto3" json:"key,omitempty"`
	UniOrBroadcast       int64    `protobuf:"varint,5,opt,name=uni_or_broadcast,json=uniOrBroadcast,proto3" json:"uni_or_broadcast,omitempty"`
	Counter              uint64   `protobuf:"varint,6,opt,name=counter,proto3" json:"counter,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Dapmsg) Reset()         { *m = Dapmsg{} }
func (m *Dapmsg) String() string { return proto.CompactTextString(m) }
func (*Dapmsg) ProtoMessage()    {}
func (*Dapmsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{0}
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

func (m *Dapmsg) GetUnixTimeReceived() int64 {
	if m != nil {
		return m.UnixTimeReceived
	}
	return 0
}

func (m *Dapmsg) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *Dapmsg) GetUniOrBroadcast() int64 {
	if m != nil {
		return m.UniOrBroadcast
	}
	return 0
}

func (m *Dapmsg) GetCounter() uint64 {
	if m != nil {
		return m.Counter
	}
	return 0
}

type GetBlockchainInfo struct {
	Version              string   `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetBlockchainInfo) Reset()         { *m = GetBlockchainInfo{} }
func (m *GetBlockchainInfo) String() string { return proto.CompactTextString(m) }
func (*GetBlockchainInfo) ProtoMessage()    {}
func (*GetBlockchainInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{1}
}
func (m *GetBlockchainInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetBlockchainInfo.Unmarshal(m, b)
}
func (m *GetBlockchainInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetBlockchainInfo.Marshal(b, m, deterministic)
}
func (dst *GetBlockchainInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlockchainInfo.Merge(dst, src)
}
func (m *GetBlockchainInfo) XXX_Size() int {
	return xxx_messageInfo_GetBlockchainInfo.Size(m)
}
func (m *GetBlockchainInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlockchainInfo.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlockchainInfo proto.InternalMessageInfo

func (m *GetBlockchainInfo) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

type ReturnBlockchainInfo struct {
	TailBlockHash        []byte   `protobuf:"bytes,1,opt,name=tail_block_hash,json=tailBlockHash,proto3" json:"tail_block_hash,omitempty"`
	BlockHeight          uint64   `protobuf:"varint,2,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	Timestamp            int64    `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReturnBlockchainInfo) Reset()         { *m = ReturnBlockchainInfo{} }
func (m *ReturnBlockchainInfo) String() string { return proto.CompactTextString(m) }
func (*ReturnBlockchainInfo) ProtoMessage()    {}
func (*ReturnBlockchainInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{2}
}
func (m *ReturnBlockchainInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReturnBlockchainInfo.Unmarshal(m, b)
}
func (m *ReturnBlockchainInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReturnBlockchainInfo.Marshal(b, m, deterministic)
}
func (dst *ReturnBlockchainInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReturnBlockchainInfo.Merge(dst, src)
}
func (m *ReturnBlockchainInfo) XXX_Size() int {
	return xxx_messageInfo_ReturnBlockchainInfo.Size(m)
}
func (m *ReturnBlockchainInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ReturnBlockchainInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ReturnBlockchainInfo proto.InternalMessageInfo

func (m *ReturnBlockchainInfo) GetTailBlockHash() []byte {
	if m != nil {
		return m.TailBlockHash
	}
	return nil
}

func (m *ReturnBlockchainInfo) GetBlockHeight() uint64 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *ReturnBlockchainInfo) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type GetBlocks struct {
	StartBlockHashes     [][]byte `protobuf:"bytes,1,rep,name=start_block_hashes,json=startBlockHashes,proto3" json:"start_block_hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetBlocks) Reset()         { *m = GetBlocks{} }
func (m *GetBlocks) String() string { return proto.CompactTextString(m) }
func (*GetBlocks) ProtoMessage()    {}
func (*GetBlocks) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{3}
}
func (m *GetBlocks) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetBlocks.Unmarshal(m, b)
}
func (m *GetBlocks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetBlocks.Marshal(b, m, deterministic)
}
func (dst *GetBlocks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlocks.Merge(dst, src)
}
func (m *GetBlocks) XXX_Size() int {
	return xxx_messageInfo_GetBlocks.Size(m)
}
func (m *GetBlocks) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlocks.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlocks proto.InternalMessageInfo

func (m *GetBlocks) GetStartBlockHashes() [][]byte {
	if m != nil {
		return m.StartBlockHashes
	}
	return nil
}

type ReturnBlocks struct {
	Blocks               []*pb.Block `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
	StartBlockHashes     [][]byte    `protobuf:"bytes,2,rep,name=start_block_hashes,json=startBlockHashes,proto3" json:"start_block_hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ReturnBlocks) Reset()         { *m = ReturnBlocks{} }
func (m *ReturnBlocks) String() string { return proto.CompactTextString(m) }
func (*ReturnBlocks) ProtoMessage()    {}
func (*ReturnBlocks) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{4}
}
func (m *ReturnBlocks) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReturnBlocks.Unmarshal(m, b)
}
func (m *ReturnBlocks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReturnBlocks.Marshal(b, m, deterministic)
}
func (dst *ReturnBlocks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReturnBlocks.Merge(dst, src)
}
func (m *ReturnBlocks) XXX_Size() int {
	return xxx_messageInfo_ReturnBlocks.Size(m)
}
func (m *ReturnBlocks) XXX_DiscardUnknown() {
	xxx_messageInfo_ReturnBlocks.DiscardUnknown(m)
}

var xxx_messageInfo_ReturnBlocks proto.InternalMessageInfo

func (m *ReturnBlocks) GetBlocks() []*pb.Block {
	if m != nil {
		return m.Blocks
	}
	return nil
}

func (m *ReturnBlocks) GetStartBlockHashes() [][]byte {
	if m != nil {
		return m.StartBlockHashes
	}
	return nil
}

type GetCommonBlocks struct {
	MsgId                int32             `protobuf:"varint,1,opt,name=msg_id,json=msgId,proto3" json:"msg_id,omitempty"`
	BlockHeaders         []*pb.BlockHeader `protobuf:"bytes,2,rep,name=block_headers,json=blockHeaders,proto3" json:"block_headers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *GetCommonBlocks) Reset()         { *m = GetCommonBlocks{} }
func (m *GetCommonBlocks) String() string { return proto.CompactTextString(m) }
func (*GetCommonBlocks) ProtoMessage()    {}
func (*GetCommonBlocks) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{5}
}
func (m *GetCommonBlocks) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCommonBlocks.Unmarshal(m, b)
}
func (m *GetCommonBlocks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCommonBlocks.Marshal(b, m, deterministic)
}
func (dst *GetCommonBlocks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCommonBlocks.Merge(dst, src)
}
func (m *GetCommonBlocks) XXX_Size() int {
	return xxx_messageInfo_GetCommonBlocks.Size(m)
}
func (m *GetCommonBlocks) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCommonBlocks.DiscardUnknown(m)
}

var xxx_messageInfo_GetCommonBlocks proto.InternalMessageInfo

func (m *GetCommonBlocks) GetMsgId() int32 {
	if m != nil {
		return m.MsgId
	}
	return 0
}

func (m *GetCommonBlocks) GetBlockHeaders() []*pb.BlockHeader {
	if m != nil {
		return m.BlockHeaders
	}
	return nil
}

type ReturnCommonBlocks struct {
	MsgId                int32             `protobuf:"varint,1,opt,name=msg_id,json=msgId,proto3" json:"msg_id,omitempty"`
	BlockHeaders         []*pb.BlockHeader `protobuf:"bytes,2,rep,name=block_headers,json=blockHeaders,proto3" json:"block_headers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ReturnCommonBlocks) Reset()         { *m = ReturnCommonBlocks{} }
func (m *ReturnCommonBlocks) String() string { return proto.CompactTextString(m) }
func (*ReturnCommonBlocks) ProtoMessage()    {}
func (*ReturnCommonBlocks) Descriptor() ([]byte, []int) {
	return fileDescriptor_dapmsg_1964a884c8452ddf, []int{6}
}
func (m *ReturnCommonBlocks) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReturnCommonBlocks.Unmarshal(m, b)
}
func (m *ReturnCommonBlocks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReturnCommonBlocks.Marshal(b, m, deterministic)
}
func (dst *ReturnCommonBlocks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReturnCommonBlocks.Merge(dst, src)
}
func (m *ReturnCommonBlocks) XXX_Size() int {
	return xxx_messageInfo_ReturnCommonBlocks.Size(m)
}
func (m *ReturnCommonBlocks) XXX_DiscardUnknown() {
	xxx_messageInfo_ReturnCommonBlocks.DiscardUnknown(m)
}

var xxx_messageInfo_ReturnCommonBlocks proto.InternalMessageInfo

func (m *ReturnCommonBlocks) GetMsgId() int32 {
	if m != nil {
		return m.MsgId
	}
	return 0
}

func (m *ReturnCommonBlocks) GetBlockHeaders() []*pb.BlockHeader {
	if m != nil {
		return m.BlockHeaders
	}
	return nil
}

func init() {
	proto.RegisterType((*Dapmsg)(nil), "networkpb.Dapmsg")
	proto.RegisterType((*GetBlockchainInfo)(nil), "networkpb.GetBlockchainInfo")
	proto.RegisterType((*ReturnBlockchainInfo)(nil), "networkpb.ReturnBlockchainInfo")
	proto.RegisterType((*GetBlocks)(nil), "networkpb.GetBlocks")
	proto.RegisterType((*ReturnBlocks)(nil), "networkpb.ReturnBlocks")
	proto.RegisterType((*GetCommonBlocks)(nil), "networkpb.GetCommonBlocks")
	proto.RegisterType((*ReturnCommonBlocks)(nil), "networkpb.ReturnCommonBlocks")
}

func init() {
	proto.RegisterFile("github.com/dappley/go-dappley/network/pb/dapmsg.proto", fileDescriptor_dapmsg_1964a884c8452ddf)
}

var fileDescriptor_dapmsg_1964a884c8452ddf = []byte{
	// 435 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xb4, 0x52, 0xc1, 0x8e, 0xd3, 0x30,
	0x10, 0x55, 0xb7, 0x6d, 0x50, 0xa6, 0x29, 0x5b, 0x0c, 0x48, 0x11, 0xe2, 0x00, 0x91, 0x40, 0x7b,
	0x60, 0x53, 0xb1, 0x08, 0x09, 0xae, 0x0b, 0x12, 0xdd, 0x13, 0x92, 0xc5, 0x3d, 0xb2, 0x9d, 0x21,
	0xb1, 0xda, 0xc4, 0x91, 0xed, 0x2c, 0xec, 0x89, 0x7f, 0xe2, 0x0b, 0xb1, 0x9d, 0x84, 0x5d, 0x90,
	0xca, 0x8d, 0xdb, 0xcc, 0x7b, 0xf3, 0xe6, 0x4d, 0x5e, 0x0c, 0x6f, 0x2b, 0x69, 0xeb, 0x9e, 0xe7,
	0x42, 0x35, 0xdb, 0x92, 0x75, 0xdd, 0x01, 0x6f, 0xb6, 0x95, 0x3a, 0x9f, 0xca, 0x16, 0xed, 0x37,
	0xa5, 0xf7, 0xdb, 0x8e, 0x7b, 0xb6, 0x31, 0x55, 0xde, 0x69, 0x65, 0x15, 0x89, 0x47, 0xa2, 0xe3,
	0x4f, 0x5e, 0xff, 0x7b, 0x83, 0x50, 0x1a, 0xbd, 0x9c, 0x1f, 0x94, 0xd8, 0x0f, 0xea, 0xec, 0xe7,
	0x0c, 0xa2, 0x8f, 0x61, 0x1d, 0xd9, 0xc0, 0x5c, 0x34, 0x65, 0x3a, 0x7b, 0x36, 0x3b, 0x8b, 0xa9,
	0x2f, 0x09, 0x81, 0x45, 0xc9, 0x2c, 0x4b, 0x4f, 0x1c, 0x94, 0xd0, 0x50, 0x93, 0x57, 0x40, 0xfa,
	0x56, 0x7e, 0x2f, 0xac, 0x6c, 0xb0, 0xd0, 0x28, 0x50, 0x5e, 0x63, 0x99, 0xce, 0xdd, 0xc4, 0x9c,
	0x6e, 0x3c, 0xf3, 0xc5, 0x11, 0x74, 0xc4, 0xfd, 0xce, 0x3d, 0xde, 0xa4, 0x8b, 0x61, 0xa7, 0x2b,
	0xc9, 0x19, 0xf8, 0xa9, 0x42, 0xe9, 0x82, 0x6b, 0xc5, 0x4a, 0xc1, 0x8c, 0x4d, 0x97, 0x41, 0x7d,
	0xdf, 0xe1, 0x9f, 0xf5, 0xe5, 0x84, 0x92, 0x14, 0xee, 0x09, 0xd5, 0xb7, 0x16, 0x75, 0x1a, 0xb9,
	0x81, 0x05, 0x9d, 0xda, 0xec, 0x1c, 0x1e, 0x7c, 0x42, 0x7b, 0xe9, 0x3f, 0x43, 0xd4, 0x4c, 0xb6,
	0x57, 0xed, 0x57, 0xe5, 0xc7, 0xaf, 0x51, 0x1b, 0xa9, 0xda, 0xf1, 0x13, 0xa6, 0x36, 0xfb, 0x01,
	0x8f, 0x28, 0xda, 0x5e, 0xb7, 0x7f, 0x29, 0x5e, 0xc2, 0xa9, 0x65, 0xf2, 0x50, 0x84, 0x3c, 0x8a,
	0x9a, 0x99, 0x3a, 0x28, 0x13, 0xba, 0xf6, 0x70, 0x18, 0xde, 0x39, 0x90, 0x3c, 0x87, 0x64, 0x1c,
	0x41, 0x59, 0xd5, 0x36, 0xc4, 0xb1, 0xa0, 0xab, 0x80, 0xed, 0x02, 0x44, 0x9e, 0x42, 0xec, 0x03,
	0x31, 0x96, 0x35, 0xdd, 0x18, 0xc6, 0x2d, 0x90, 0xbd, 0x87, 0x78, 0xba, 0xd7, 0xf8, 0x00, 0x1d,
	0xaa, 0xed, 0x1d, 0x5b, 0x34, 0xce, 0x78, 0xee, 0x8c, 0x37, 0x81, 0xf9, 0xed, 0x8c, 0x26, 0x13,
	0x90, 0xdc, 0xb9, 0xdd, 0x90, 0x17, 0x10, 0x05, 0xdd, 0xa0, 0x58, 0x5d, 0xac, 0x73, 0xff, 0x57,
	0x3b, 0x9e, 0x07, 0x9e, 0x8e, 0xe4, 0x11, 0x93, 0x93, 0x23, 0x26, 0x1c, 0x4e, 0xdd, 0x7d, 0x1f,
	0x54, 0xd3, 0xa8, 0xc9, 0xe7, 0x31, 0x44, 0xee, 0x4d, 0x14, 0x72, 0x78, 0x0f, 0x4b, 0xba, 0x74,
	0xdd, 0x55, 0x49, 0xde, 0xc1, 0x7a, 0x8a, 0x82, 0x95, 0x2e, 0xde, 0xb0, 0x72, 0x75, 0xf1, 0xf0,
	0x8f, 0x2b, 0x76, 0x81, 0xa3, 0x09, 0xbf, 0x6d, 0x4c, 0x86, 0x40, 0x86, 0x0f, 0xf9, 0xaf, 0x36,
	0x3c, 0x0a, 0xcf, 0xfa, 0xcd, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x2d, 0x6e, 0x1f, 0x4d, 0x4d,
	0x03, 0x00, 0x00,
}
