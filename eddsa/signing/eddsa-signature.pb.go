// Copyright © 2019 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: protob/eddsa-signature.proto

package signing

import (
	common "github.com/binance-chain/tss-lib/common"
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

//
// State object for signatures, contains the final EdDSA signature.
type SignatureData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signature *common.ECSignature `protobuf:"bytes,10,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *SignatureData) Reset() {
	*x = SignatureData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protob_eddsa_signature_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignatureData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignatureData) ProtoMessage() {}

func (x *SignatureData) ProtoReflect() protoreflect.Message {
	mi := &file_protob_eddsa_signature_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignatureData.ProtoReflect.Descriptor instead.
func (*SignatureData) Descriptor() ([]byte, []int) {
	return file_protob_eddsa_signature_proto_rawDescGZIP(), []int{0}
}

func (x *SignatureData) GetSignature() *common.ECSignature {
	if x != nil {
		return x.Signature
	}
	return nil
}

var File_protob_eddsa_signature_proto protoreflect.FileDescriptor

var file_protob_eddsa_signature_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x2f, 0x65, 0x64, 0x64, 0x73, 0x61, 0x2d, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x2f, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x3b, 0x0a, 0x0d, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x2a, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x45, 0x43, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62,
	0x69, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2f, 0x74, 0x73, 0x73,
	0x2d, 0x6c, 0x69, 0x62, 0x2f, 0x65, 0x64, 0x64, 0x73, 0x61, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x69,
	0x6e, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protob_eddsa_signature_proto_rawDescOnce sync.Once
	file_protob_eddsa_signature_proto_rawDescData = file_protob_eddsa_signature_proto_rawDesc
)

func file_protob_eddsa_signature_proto_rawDescGZIP() []byte {
	file_protob_eddsa_signature_proto_rawDescOnce.Do(func() {
		file_protob_eddsa_signature_proto_rawDescData = protoimpl.X.CompressGZIP(file_protob_eddsa_signature_proto_rawDescData)
	})
	return file_protob_eddsa_signature_proto_rawDescData
}

var file_protob_eddsa_signature_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protob_eddsa_signature_proto_goTypes = []interface{}{
	(*SignatureData)(nil),      // 0: SignatureData
	(*common.ECSignature)(nil), // 1: ECSignature
}
var file_protob_eddsa_signature_proto_depIdxs = []int32{
	1, // 0: SignatureData.signature:type_name -> ECSignature
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_protob_eddsa_signature_proto_init() }
func file_protob_eddsa_signature_proto_init() {
	if File_protob_eddsa_signature_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protob_eddsa_signature_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignatureData); i {
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
			RawDescriptor: file_protob_eddsa_signature_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protob_eddsa_signature_proto_goTypes,
		DependencyIndexes: file_protob_eddsa_signature_proto_depIdxs,
		MessageInfos:      file_protob_eddsa_signature_proto_msgTypes,
	}.Build()
	File_protob_eddsa_signature_proto = out.File
	file_protob_eddsa_signature_proto_rawDesc = nil
	file_protob_eddsa_signature_proto_goTypes = nil
	file_protob_eddsa_signature_proto_depIdxs = nil
}
