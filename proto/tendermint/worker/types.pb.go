// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tendermint/worker/types.proto

package worker

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/cosmos/gogoproto/types"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type BatchData struct {
	Txs          [][]byte  `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
	BatchKey     []byte    `protobuf:"bytes,2,opt,name=batch_key,json=batchKey,proto3" json:"batch_key,omitempty"`
	BatchOwner   string    `protobuf:"bytes,3,opt,name=batch_owner,json=batchOwner,proto3" json:"batch_owner,omitempty"`
	BatchCreated time.Time `protobuf:"bytes,5,opt,name=batch_created,json=batchCreated,proto3,stdtime" json:"batch_created"`
}

func (m *BatchData) Reset()         { *m = BatchData{} }
func (m *BatchData) String() string { return proto.CompactTextString(m) }
func (*BatchData) ProtoMessage()    {}
func (*BatchData) Descriptor() ([]byte, []int) {
	return fileDescriptor_5b155f9c34ea770a, []int{0}
}
func (m *BatchData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BatchData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BatchData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BatchData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BatchData.Merge(m, src)
}
func (m *BatchData) XXX_Size() int {
	return m.Size()
}
func (m *BatchData) XXX_DiscardUnknown() {
	xxx_messageInfo_BatchData.DiscardUnknown(m)
}

var xxx_messageInfo_BatchData proto.InternalMessageInfo

func (m *BatchData) GetTxs() [][]byte {
	if m != nil {
		return m.Txs
	}
	return nil
}

func (m *BatchData) GetBatchKey() []byte {
	if m != nil {
		return m.BatchKey
	}
	return nil
}

func (m *BatchData) GetBatchOwner() string {
	if m != nil {
		return m.BatchOwner
	}
	return ""
}

func (m *BatchData) GetBatchCreated() time.Time {
	if m != nil {
		return m.BatchCreated
	}
	return time.Time{}
}

type BatchAck struct {
	BatchKey    []byte    `protobuf:"bytes,1,opt,name=batch_key,json=batchKey,proto3" json:"batch_key,omitempty"`
	BatchOwner  string    `protobuf:"bytes,2,opt,name=batch_owner,json=batchOwner,proto3" json:"batch_owner,omitempty"`
	BatchRecved time.Time `protobuf:"bytes,3,opt,name=batch_recved,json=batchRecved,proto3,stdtime" json:"batch_recved"`
}

func (m *BatchAck) Reset()         { *m = BatchAck{} }
func (m *BatchAck) String() string { return proto.CompactTextString(m) }
func (*BatchAck) ProtoMessage()    {}
func (*BatchAck) Descriptor() ([]byte, []int) {
	return fileDescriptor_5b155f9c34ea770a, []int{1}
}
func (m *BatchAck) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BatchAck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BatchAck.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BatchAck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BatchAck.Merge(m, src)
}
func (m *BatchAck) XXX_Size() int {
	return m.Size()
}
func (m *BatchAck) XXX_DiscardUnknown() {
	xxx_messageInfo_BatchAck.DiscardUnknown(m)
}

var xxx_messageInfo_BatchAck proto.InternalMessageInfo

func (m *BatchAck) GetBatchKey() []byte {
	if m != nil {
		return m.BatchKey
	}
	return nil
}

func (m *BatchAck) GetBatchOwner() string {
	if m != nil {
		return m.BatchOwner
	}
	return ""
}

func (m *BatchAck) GetBatchRecved() time.Time {
	if m != nil {
		return m.BatchRecved
	}
	return time.Time{}
}

type Message struct {
	// Types that are valid to be assigned to Sum:
	//
	//	*Message_BatchData
	//	*Message_BatchAck
	Sum isMessage_Sum `protobuf_oneof:"sum"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_5b155f9c34ea770a, []int{2}
}
func (m *Message) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Message.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return m.Size()
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

type isMessage_Sum interface {
	isMessage_Sum()
	MarshalTo([]byte) (int, error)
	Size() int
}

type Message_BatchData struct {
	BatchData *BatchData `protobuf:"bytes,1,opt,name=batch_data,json=batchData,proto3,oneof" json:"batch_data,omitempty"`
}
type Message_BatchAck struct {
	BatchAck *BatchAck `protobuf:"bytes,2,opt,name=batch_ack,json=batchAck,proto3,oneof" json:"batch_ack,omitempty"`
}

func (*Message_BatchData) isMessage_Sum() {}
func (*Message_BatchAck) isMessage_Sum()  {}

func (m *Message) GetSum() isMessage_Sum {
	if m != nil {
		return m.Sum
	}
	return nil
}

func (m *Message) GetBatchData() *BatchData {
	if x, ok := m.GetSum().(*Message_BatchData); ok {
		return x.BatchData
	}
	return nil
}

func (m *Message) GetBatchAck() *BatchAck {
	if x, ok := m.GetSum().(*Message_BatchAck); ok {
		return x.BatchAck
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Message) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Message_BatchData)(nil),
		(*Message_BatchAck)(nil),
	}
}

func init() {
	proto.RegisterType((*BatchData)(nil), "tendermint.worker.BatchData")
	proto.RegisterType((*BatchAck)(nil), "tendermint.worker.BatchAck")
	proto.RegisterType((*Message)(nil), "tendermint.worker.Message")
}

func init() { proto.RegisterFile("tendermint/worker/types.proto", fileDescriptor_5b155f9c34ea770a) }

var fileDescriptor_5b155f9c34ea770a = []byte{
	// 393 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xc1, 0x6e, 0xda, 0x40,
	0x10, 0x86, 0xbd, 0x58, 0xb4, 0xb0, 0xa6, 0x52, 0x6b, 0xf5, 0x60, 0x41, 0x6b, 0x5b, 0x9c, 0x7c,
	0x5a, 0x4b, 0xb4, 0xa7, 0x4a, 0x3d, 0xe0, 0x56, 0x0a, 0x51, 0x94, 0x44, 0xb2, 0x72, 0xca, 0x05,
	0xad, 0xd7, 0x83, 0x41, 0x8e, 0x59, 0x64, 0x2f, 0x21, 0x3c, 0x43, 0x2e, 0xe4, 0x25, 0xf2, 0x2c,
	0x1c, 0x39, 0xe6, 0x94, 0x44, 0xf0, 0x22, 0x91, 0x77, 0x43, 0x50, 0x84, 0x12, 0xe5, 0x36, 0x9e,
	0x7f, 0xfe, 0xf1, 0x37, 0xb3, 0x83, 0x7f, 0x0a, 0x18, 0xc7, 0x90, 0x67, 0xa3, 0xb1, 0xf0, 0x67,
	0x3c, 0x4f, 0x21, 0xf7, 0xc5, 0x7c, 0x02, 0x05, 0x99, 0xe4, 0x5c, 0x70, 0xf3, 0xdb, 0x4e, 0x26,
	0x4a, 0x6e, 0x7e, 0x4f, 0x78, 0xc2, 0xa5, 0xea, 0x97, 0x91, 0x2a, 0x6c, 0x3a, 0x09, 0xe7, 0xc9,
	0x05, 0xf8, 0xf2, 0x2b, 0x9a, 0x0e, 0x7c, 0x31, 0xca, 0xa0, 0x10, 0x34, 0x9b, 0xa8, 0x82, 0xf6,
	0x2d, 0xc2, 0xf5, 0x80, 0x0a, 0x36, 0xfc, 0x4f, 0x05, 0x35, 0xbf, 0x62, 0x5d, 0x5c, 0x15, 0x16,
	0x72, 0x75, 0xaf, 0x11, 0x96, 0xa1, 0xd9, 0xc2, 0xf5, 0xa8, 0x94, 0xfb, 0x29, 0xcc, 0xad, 0x8a,
	0x8b, 0xbc, 0x46, 0x58, 0x93, 0x89, 0x23, 0x98, 0x9b, 0x0e, 0x36, 0x94, 0xc8, 0x67, 0x63, 0xc8,
	0x2d, 0xdd, 0x45, 0x5e, 0x3d, 0xc4, 0x32, 0x75, 0x5a, 0x66, 0xcc, 0x43, 0xfc, 0x45, 0x15, 0xb0,
	0x1c, 0xa8, 0x80, 0xd8, 0xaa, 0xba, 0xc8, 0x33, 0x3a, 0x4d, 0xa2, 0xb0, 0xc8, 0x16, 0x8b, 0x9c,
	0x6d, 0xb1, 0x82, 0xda, 0xf2, 0xde, 0xd1, 0x16, 0x0f, 0x0e, 0x0a, 0x1b, 0xd2, 0xfa, 0x4f, 0x39,
	0xdb, 0x37, 0x08, 0xd7, 0x24, 0x68, 0x97, 0xa5, 0xaf, 0xa9, 0xd0, 0xfb, 0x54, 0x95, 0x3d, 0xaa,
	0x03, 0xac, 0x5a, 0xf7, 0x73, 0x60, 0x97, 0x10, 0x4b, 0xee, 0x8f, 0x42, 0xa9, 0xd6, 0xa1, 0x34,
	0xb6, 0xaf, 0x11, 0xfe, 0x7c, 0x0c, 0x45, 0x41, 0x13, 0x30, 0xff, 0x62, 0xf5, 0x8b, 0x7e, 0x4c,
	0x05, 0x95, 0x4c, 0x46, 0xe7, 0x07, 0xd9, 0x7b, 0x27, 0xf2, 0xb2, 0xec, 0x9e, 0x16, 0xaa, 0x21,
	0xe4, 0xe6, 0xff, 0x6c, 0x27, 0xa2, 0x2c, 0x95, 0xc8, 0x46, 0xa7, 0xf5, 0x96, 0xbb, 0xcb, 0xd2,
	0x9e, 0xf6, 0x3c, 0x70, 0x97, 0xa5, 0x41, 0x15, 0xeb, 0xc5, 0x34, 0x0b, 0x4e, 0x96, 0x6b, 0x1b,
	0xad, 0xd6, 0x36, 0x7a, 0x5c, 0xdb, 0x68, 0xb1, 0xb1, 0xb5, 0xd5, 0xc6, 0xd6, 0xee, 0x36, 0xb6,
	0x76, 0xfe, 0x3b, 0x19, 0x89, 0xe1, 0x34, 0x22, 0x8c, 0x67, 0x3e, 0xe3, 0x19, 0x88, 0x68, 0x20,
	0x76, 0x81, 0x3a, 0x9a, 0xbd, 0x83, 0x8b, 0x3e, 0x49, 0xe1, 0xd7, 0x53, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xc3, 0x39, 0xeb, 0x35, 0x8c, 0x02, 0x00, 0x00,
}

func (m *BatchData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BatchData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BatchData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.BatchCreated, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.BatchCreated):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintTypes(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x2a
	if len(m.BatchOwner) > 0 {
		i -= len(m.BatchOwner)
		copy(dAtA[i:], m.BatchOwner)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.BatchOwner)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.BatchKey) > 0 {
		i -= len(m.BatchKey)
		copy(dAtA[i:], m.BatchKey)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.BatchKey)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Txs) > 0 {
		for iNdEx := len(m.Txs) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Txs[iNdEx])
			copy(dAtA[i:], m.Txs[iNdEx])
			i = encodeVarintTypes(dAtA, i, uint64(len(m.Txs[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *BatchAck) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BatchAck) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BatchAck) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.BatchRecved, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.BatchRecved):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintTypes(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x1a
	if len(m.BatchOwner) > 0 {
		i -= len(m.BatchOwner)
		copy(dAtA[i:], m.BatchOwner)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.BatchOwner)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.BatchKey) > 0 {
		i -= len(m.BatchKey)
		copy(dAtA[i:], m.BatchKey)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.BatchKey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Message) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Message) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Sum != nil {
		{
			size := m.Sum.Size()
			i -= size
			if _, err := m.Sum.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	return len(dAtA) - i, nil
}

func (m *Message_BatchData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_BatchData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.BatchData != nil {
		{
			size, err := m.BatchData.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}
func (m *Message_BatchAck) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Message_BatchAck) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.BatchAck != nil {
		{
			size, err := m.BatchAck.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	return len(dAtA) - i, nil
}
func encodeVarintTypes(dAtA []byte, offset int, v uint64) int {
	offset -= sovTypes(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *BatchData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Txs) > 0 {
		for _, b := range m.Txs {
			l = len(b)
			n += 1 + l + sovTypes(uint64(l))
		}
	}
	l = len(m.BatchKey)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.BatchOwner)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.BatchCreated)
	n += 1 + l + sovTypes(uint64(l))
	return n
}

func (m *BatchAck) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BatchKey)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.BatchOwner)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.BatchRecved)
	n += 1 + l + sovTypes(uint64(l))
	return n
}

func (m *Message) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Sum != nil {
		n += m.Sum.Size()
	}
	return n
}

func (m *Message_BatchData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BatchData != nil {
		l = m.BatchData.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}
func (m *Message_BatchAck) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BatchAck != nil {
		l = m.BatchAck.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func sovTypes(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTypes(x uint64) (n int) {
	return sovTypes(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *BatchData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BatchData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BatchData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Txs", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Txs = append(m.Txs, make([]byte, postIndex-iNdEx))
			copy(m.Txs[len(m.Txs)-1], dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BatchKey = append(m.BatchKey[:0], dAtA[iNdEx:postIndex]...)
			if m.BatchKey == nil {
				m.BatchKey = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchOwner", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BatchOwner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchCreated", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.BatchCreated, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BatchAck) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BatchAck: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BatchAck: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BatchKey = append(m.BatchKey[:0], dAtA[iNdEx:postIndex]...)
			if m.BatchKey == nil {
				m.BatchKey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchOwner", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BatchOwner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchRecved", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.BatchRecved, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Message) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Message: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Message: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchData", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &BatchData{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_BatchData{v}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BatchAck", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &BatchAck{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Sum = &Message_BatchAck{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTypes(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTypes
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTypes
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTypes
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTypes        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTypes          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTypes = fmt.Errorf("proto: unexpected end of group")
)
