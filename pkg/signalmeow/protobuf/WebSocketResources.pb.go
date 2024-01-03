//*
// Copyright (C) 2014-2016 Open Whisper Systems
//
// Licensed according to the LICENSE file in this repository.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.12
// source: WebSocketResources.proto

package signalpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type WebSocketMessage_Type int32

const (
	WebSocketMessage_UNKNOWN  WebSocketMessage_Type = 0
	WebSocketMessage_REQUEST  WebSocketMessage_Type = 1
	WebSocketMessage_RESPONSE WebSocketMessage_Type = 2
)

// Enum value maps for WebSocketMessage_Type.
var (
	WebSocketMessage_Type_name = map[int32]string{
		0: "UNKNOWN",
		1: "REQUEST",
		2: "RESPONSE",
	}
	WebSocketMessage_Type_value = map[string]int32{
		"UNKNOWN":  0,
		"REQUEST":  1,
		"RESPONSE": 2,
	}
)

func (x WebSocketMessage_Type) Enum() *WebSocketMessage_Type {
	p := new(WebSocketMessage_Type)
	*p = x
	return p
}

func (x WebSocketMessage_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (WebSocketMessage_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_WebSocketResources_proto_enumTypes[0].Descriptor()
}

func (WebSocketMessage_Type) Type() protoreflect.EnumType {
	return &file_WebSocketResources_proto_enumTypes[0]
}

func (x WebSocketMessage_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *WebSocketMessage_Type) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = WebSocketMessage_Type(num)
	return nil
}

// Deprecated: Use WebSocketMessage_Type.Descriptor instead.
func (WebSocketMessage_Type) EnumDescriptor() ([]byte, []int) {
	return file_WebSocketResources_proto_rawDescGZIP(), []int{2, 0}
}

type WebSocketRequestMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Verb    *string  `protobuf:"bytes,1,opt,name=verb" json:"verb,omitempty"`
	Path    *string  `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	Body    []byte   `protobuf:"bytes,3,opt,name=body" json:"body,omitempty"`
	Headers []string `protobuf:"bytes,5,rep,name=headers" json:"headers,omitempty"`
	Id      *uint64  `protobuf:"varint,4,opt,name=id" json:"id,omitempty"`
}

func (x *WebSocketRequestMessage) Reset() {
	*x = WebSocketRequestMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_WebSocketResources_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebSocketRequestMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebSocketRequestMessage) ProtoMessage() {}

func (x *WebSocketRequestMessage) ProtoReflect() protoreflect.Message {
	mi := &file_WebSocketResources_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebSocketRequestMessage.ProtoReflect.Descriptor instead.
func (*WebSocketRequestMessage) Descriptor() ([]byte, []int) {
	return file_WebSocketResources_proto_rawDescGZIP(), []int{0}
}

func (x *WebSocketRequestMessage) GetVerb() string {
	if x != nil && x.Verb != nil {
		return *x.Verb
	}
	return ""
}

func (x *WebSocketRequestMessage) GetPath() string {
	if x != nil && x.Path != nil {
		return *x.Path
	}
	return ""
}

func (x *WebSocketRequestMessage) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *WebSocketRequestMessage) GetHeaders() []string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *WebSocketRequestMessage) GetId() uint64 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

type WebSocketResponseMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      *uint64  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Status  *uint32  `protobuf:"varint,2,opt,name=status" json:"status,omitempty"`
	Message *string  `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
	Headers []string `protobuf:"bytes,5,rep,name=headers" json:"headers,omitempty"`
	Body    []byte   `protobuf:"bytes,4,opt,name=body" json:"body,omitempty"`
}

func (x *WebSocketResponseMessage) Reset() {
	*x = WebSocketResponseMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_WebSocketResources_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebSocketResponseMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebSocketResponseMessage) ProtoMessage() {}

func (x *WebSocketResponseMessage) ProtoReflect() protoreflect.Message {
	mi := &file_WebSocketResources_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebSocketResponseMessage.ProtoReflect.Descriptor instead.
func (*WebSocketResponseMessage) Descriptor() ([]byte, []int) {
	return file_WebSocketResources_proto_rawDescGZIP(), []int{1}
}

func (x *WebSocketResponseMessage) GetId() uint64 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *WebSocketResponseMessage) GetStatus() uint32 {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return 0
}

func (x *WebSocketResponseMessage) GetMessage() string {
	if x != nil && x.Message != nil {
		return *x.Message
	}
	return ""
}

func (x *WebSocketResponseMessage) GetHeaders() []string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *WebSocketResponseMessage) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

type WebSocketMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     *WebSocketMessage_Type    `protobuf:"varint,1,opt,name=type,enum=signalservice.WebSocketMessage_Type" json:"type,omitempty"`
	Request  *WebSocketRequestMessage  `protobuf:"bytes,2,opt,name=request" json:"request,omitempty"`
	Response *WebSocketResponseMessage `protobuf:"bytes,3,opt,name=response" json:"response,omitempty"`
}

func (x *WebSocketMessage) Reset() {
	*x = WebSocketMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_WebSocketResources_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebSocketMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebSocketMessage) ProtoMessage() {}

func (x *WebSocketMessage) ProtoReflect() protoreflect.Message {
	mi := &file_WebSocketResources_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebSocketMessage.ProtoReflect.Descriptor instead.
func (*WebSocketMessage) Descriptor() ([]byte, []int) {
	return file_WebSocketResources_proto_rawDescGZIP(), []int{2}
}

func (x *WebSocketMessage) GetType() WebSocketMessage_Type {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return WebSocketMessage_UNKNOWN
}

func (x *WebSocketMessage) GetRequest() *WebSocketRequestMessage {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *WebSocketMessage) GetResponse() *WebSocketResponseMessage {
	if x != nil {
		return x.Response
	}
	return nil
}

var File_WebSocketResources_proto protoreflect.FileDescriptor

var file_WebSocketResources_proto_rawDesc = []byte{
	0x0a, 0x18, 0x57, 0x65, 0x62, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x6c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x22, 0x7f, 0x0a, 0x17, 0x57, 0x65, 0x62,
	0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x65, 0x72, 0x62, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x76, 0x65, 0x72, 0x62, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04,
	0x62, 0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79,
	0x12, 0x18, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x22, 0x8a, 0x01, 0x0a, 0x18, 0x57,
	0x65, 0x62, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x22, 0x83, 0x02, 0x0a, 0x10, 0x57, 0x65, 0x62, 0x53,
	0x6f, 0x63, 0x6b, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x38, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x6c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x57, 0x65, 0x62, 0x53, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x40, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x57, 0x65, 0x62, 0x53, 0x6f, 0x63, 0x6b, 0x65,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52,
	0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x43, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x6c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x57, 0x65, 0x62, 0x53, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2e, 0x0a,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e,
	0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x01, 0x12,
	0x0c, 0x0a, 0x08, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x10, 0x02, 0x42, 0x46, 0x0a,
	0x33, 0x6f, 0x72, 0x67, 0x2e, 0x77, 0x68, 0x69, 0x73, 0x70, 0x65, 0x72, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x73, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x77, 0x65, 0x62, 0x73, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x42, 0x0f, 0x57, 0x65, 0x62, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x73,
}

var (
	file_WebSocketResources_proto_rawDescOnce sync.Once
	file_WebSocketResources_proto_rawDescData = file_WebSocketResources_proto_rawDesc
)

func file_WebSocketResources_proto_rawDescGZIP() []byte {
	file_WebSocketResources_proto_rawDescOnce.Do(func() {
		file_WebSocketResources_proto_rawDescData = protoimpl.X.CompressGZIP(file_WebSocketResources_proto_rawDescData)
	})
	return file_WebSocketResources_proto_rawDescData
}

var file_WebSocketResources_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_WebSocketResources_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_WebSocketResources_proto_goTypes = []interface{}{
	(WebSocketMessage_Type)(0),       // 0: signalservice.WebSocketMessage.Type
	(*WebSocketRequestMessage)(nil),  // 1: signalservice.WebSocketRequestMessage
	(*WebSocketResponseMessage)(nil), // 2: signalservice.WebSocketResponseMessage
	(*WebSocketMessage)(nil),         // 3: signalservice.WebSocketMessage
}
var file_WebSocketResources_proto_depIdxs = []int32{
	0, // 0: signalservice.WebSocketMessage.type:type_name -> signalservice.WebSocketMessage.Type
	1, // 1: signalservice.WebSocketMessage.request:type_name -> signalservice.WebSocketRequestMessage
	2, // 2: signalservice.WebSocketMessage.response:type_name -> signalservice.WebSocketResponseMessage
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_WebSocketResources_proto_init() }
func file_WebSocketResources_proto_init() {
	if File_WebSocketResources_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_WebSocketResources_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebSocketRequestMessage); i {
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
		file_WebSocketResources_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebSocketResponseMessage); i {
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
		file_WebSocketResources_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebSocketMessage); i {
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
			RawDescriptor: file_WebSocketResources_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_WebSocketResources_proto_goTypes,
		DependencyIndexes: file_WebSocketResources_proto_depIdxs,
		EnumInfos:         file_WebSocketResources_proto_enumTypes,
		MessageInfos:      file_WebSocketResources_proto_msgTypes,
	}.Build()
	File_WebSocketResources_proto = out.File
	file_WebSocketResources_proto_rawDesc = nil
	file_WebSocketResources_proto_goTypes = nil
	file_WebSocketResources_proto_depIdxs = nil
}
