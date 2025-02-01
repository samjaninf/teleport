// Copyright 2024 Gravitational, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: teleport/provisioning/v1/provisioning_service.proto

package provisioningv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// DeleteDownstreamProvisioningStatesRequest is a request to delete all provisioning states for
// a given DownstreamId.
type DeleteDownstreamProvisioningStatesRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// DownstreamId identifies the downstream service that this state applies to.
	DownstreamId  string `protobuf:"bytes,1,opt,name=downstream_id,json=downstreamId,proto3" json:"downstream_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteDownstreamProvisioningStatesRequest) Reset() {
	*x = DeleteDownstreamProvisioningStatesRequest{}
	mi := &file_teleport_provisioning_v1_provisioning_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteDownstreamProvisioningStatesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteDownstreamProvisioningStatesRequest) ProtoMessage() {}

func (x *DeleteDownstreamProvisioningStatesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_provisioning_v1_provisioning_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteDownstreamProvisioningStatesRequest.ProtoReflect.Descriptor instead.
func (*DeleteDownstreamProvisioningStatesRequest) Descriptor() ([]byte, []int) {
	return file_teleport_provisioning_v1_provisioning_service_proto_rawDescGZIP(), []int{0}
}

func (x *DeleteDownstreamProvisioningStatesRequest) GetDownstreamId() string {
	if x != nil {
		return x.DownstreamId
	}
	return ""
}

var File_teleport_provisioning_v1_provisioning_service_proto protoreflect.FileDescriptor

var file_teleport_provisioning_v1_provisioning_service_proto_rawDesc = string([]byte{
	0x0a, 0x33, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a,
	0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x50, 0x0a, 0x29,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x44, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x6f, 0x77,
	0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x64, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x32, 0x99,
	0x01, 0x0a, 0x13, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x81, 0x01, 0x0a, 0x22, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x44, 0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x50, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x65, 0x73, 0x12, 0x43, 0x2e,
	0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x44,
	0x6f, 0x77, 0x6e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x5c, 0x5a, 0x5a, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x72, 0x61, 0x76, 0x69, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f,
	0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x3b, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_teleport_provisioning_v1_provisioning_service_proto_rawDescOnce sync.Once
	file_teleport_provisioning_v1_provisioning_service_proto_rawDescData []byte
)

func file_teleport_provisioning_v1_provisioning_service_proto_rawDescGZIP() []byte {
	file_teleport_provisioning_v1_provisioning_service_proto_rawDescOnce.Do(func() {
		file_teleport_provisioning_v1_provisioning_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_teleport_provisioning_v1_provisioning_service_proto_rawDesc), len(file_teleport_provisioning_v1_provisioning_service_proto_rawDesc)))
	})
	return file_teleport_provisioning_v1_provisioning_service_proto_rawDescData
}

var file_teleport_provisioning_v1_provisioning_service_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_teleport_provisioning_v1_provisioning_service_proto_goTypes = []any{
	(*DeleteDownstreamProvisioningStatesRequest)(nil), // 0: teleport.provisioning.v1.DeleteDownstreamProvisioningStatesRequest
	(*emptypb.Empty)(nil),                             // 1: google.protobuf.Empty
}
var file_teleport_provisioning_v1_provisioning_service_proto_depIdxs = []int32{
	0, // 0: teleport.provisioning.v1.ProvisioningService.DeleteDownstreamProvisioningStates:input_type -> teleport.provisioning.v1.DeleteDownstreamProvisioningStatesRequest
	1, // 1: teleport.provisioning.v1.ProvisioningService.DeleteDownstreamProvisioningStates:output_type -> google.protobuf.Empty
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_teleport_provisioning_v1_provisioning_service_proto_init() }
func file_teleport_provisioning_v1_provisioning_service_proto_init() {
	if File_teleport_provisioning_v1_provisioning_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_teleport_provisioning_v1_provisioning_service_proto_rawDesc), len(file_teleport_provisioning_v1_provisioning_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_teleport_provisioning_v1_provisioning_service_proto_goTypes,
		DependencyIndexes: file_teleport_provisioning_v1_provisioning_service_proto_depIdxs,
		MessageInfos:      file_teleport_provisioning_v1_provisioning_service_proto_msgTypes,
	}.Build()
	File_teleport_provisioning_v1_provisioning_service_proto = out.File
	file_teleport_provisioning_v1_provisioning_service_proto_goTypes = nil
	file_teleport_provisioning_v1_provisioning_service_proto_depIdxs = nil
}
