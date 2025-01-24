// Copyright 2024 Gravitational, Inc
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
// source: teleport/machineid/v1/bot_instance_service.proto

package machineidv1

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

// Request for GetBotInstance.
type GetBotInstanceRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the bot associated with the instance.
	BotName string `protobuf:"bytes,1,opt,name=bot_name,json=botName,proto3" json:"bot_name,omitempty"`
	// The unique identifier of the bot instance to retrieve.
	InstanceId    string `protobuf:"bytes,2,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetBotInstanceRequest) Reset() {
	*x = GetBotInstanceRequest{}
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetBotInstanceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBotInstanceRequest) ProtoMessage() {}

func (x *GetBotInstanceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBotInstanceRequest.ProtoReflect.Descriptor instead.
func (*GetBotInstanceRequest) Descriptor() ([]byte, []int) {
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP(), []int{0}
}

func (x *GetBotInstanceRequest) GetBotName() string {
	if x != nil {
		return x.BotName
	}
	return ""
}

func (x *GetBotInstanceRequest) GetInstanceId() string {
	if x != nil {
		return x.InstanceId
	}
	return ""
}

// Request for ListBotInstances.
//
// Follows the pagination semantics of
// https://cloud.google.com/apis/design/standard_methods#list
type ListBotInstancesRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the Bot to list BotInstances for. If empty, all BotInstances
	// will be listed.
	FilterBotName string `protobuf:"bytes,1,opt,name=filter_bot_name,json=filterBotName,proto3" json:"filter_bot_name,omitempty"`
	// The maximum number of items to return.
	// The server may impose a different page size at its discretion.
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// The page_token value returned from a previous ListBotInstances request, if
	// any.
	PageToken     string `protobuf:"bytes,3,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListBotInstancesRequest) Reset() {
	*x = ListBotInstancesRequest{}
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListBotInstancesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListBotInstancesRequest) ProtoMessage() {}

func (x *ListBotInstancesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListBotInstancesRequest.ProtoReflect.Descriptor instead.
func (*ListBotInstancesRequest) Descriptor() ([]byte, []int) {
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP(), []int{1}
}

func (x *ListBotInstancesRequest) GetFilterBotName() string {
	if x != nil {
		return x.FilterBotName
	}
	return ""
}

func (x *ListBotInstancesRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListBotInstancesRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

// Response for ListBotInstances.
type ListBotInstancesResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// BotInstance that matched the search.
	BotInstances []*BotInstance `protobuf:"bytes,1,rep,name=bot_instances,json=botInstances,proto3" json:"bot_instances,omitempty"`
	// Token to retrieve the next page of results, or empty if there are no
	// more results exist.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListBotInstancesResponse) Reset() {
	*x = ListBotInstancesResponse{}
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListBotInstancesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListBotInstancesResponse) ProtoMessage() {}

func (x *ListBotInstancesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListBotInstancesResponse.ProtoReflect.Descriptor instead.
func (*ListBotInstancesResponse) Descriptor() ([]byte, []int) {
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP(), []int{2}
}

func (x *ListBotInstancesResponse) GetBotInstances() []*BotInstance {
	if x != nil {
		return x.BotInstances
	}
	return nil
}

func (x *ListBotInstancesResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

// Request for DeleteBotInstance.
type DeleteBotInstanceRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The name of the BotInstance to delete.
	BotName string `protobuf:"bytes,1,opt,name=bot_name,json=botName,proto3" json:"bot_name,omitempty"`
	// The unique identifier of the bot instance to delete.
	InstanceId    string `protobuf:"bytes,2,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteBotInstanceRequest) Reset() {
	*x = DeleteBotInstanceRequest{}
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteBotInstanceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteBotInstanceRequest) ProtoMessage() {}

func (x *DeleteBotInstanceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteBotInstanceRequest.ProtoReflect.Descriptor instead.
func (*DeleteBotInstanceRequest) Descriptor() ([]byte, []int) {
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP(), []int{3}
}

func (x *DeleteBotInstanceRequest) GetBotName() string {
	if x != nil {
		return x.BotName
	}
	return ""
}

func (x *DeleteBotInstanceRequest) GetInstanceId() string {
	if x != nil {
		return x.InstanceId
	}
	return ""
}

// The request for SubmitHeartbeat.
type SubmitHeartbeatRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The heartbeat data to submit.
	Heartbeat     *BotInstanceStatusHeartbeat `protobuf:"bytes,1,opt,name=heartbeat,proto3" json:"heartbeat,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitHeartbeatRequest) Reset() {
	*x = SubmitHeartbeatRequest{}
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitHeartbeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitHeartbeatRequest) ProtoMessage() {}

func (x *SubmitHeartbeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitHeartbeatRequest.ProtoReflect.Descriptor instead.
func (*SubmitHeartbeatRequest) Descriptor() ([]byte, []int) {
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP(), []int{4}
}

func (x *SubmitHeartbeatRequest) GetHeartbeat() *BotInstanceStatusHeartbeat {
	if x != nil {
		return x.Heartbeat
	}
	return nil
}

// The response for SubmitHeartbeat.
type SubmitHeartbeatResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitHeartbeatResponse) Reset() {
	*x = SubmitHeartbeatResponse{}
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitHeartbeatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitHeartbeatResponse) ProtoMessage() {}

func (x *SubmitHeartbeatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_machineid_v1_bot_instance_service_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitHeartbeatResponse.ProtoReflect.Descriptor instead.
func (*SubmitHeartbeatResponse) Descriptor() ([]byte, []int) {
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP(), []int{5}
}

var File_teleport_machineid_v1_bot_instance_service_proto protoreflect.FileDescriptor

var file_teleport_machineid_v1_bot_instance_service_proto_rawDesc = string([]byte{
	0x0a, 0x30, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x6d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x69, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x6f, 0x74, 0x5f, 0x69, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x15, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x28, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x2f, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x6f,
	0x74, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x53, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x62, 0x6f, 0x74,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x62, 0x6f, 0x74,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x49, 0x64, 0x22, 0x7d, 0x0a, 0x17, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x74,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x26, 0x0a, 0x0f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x5f, 0x62, 0x6f, 0x74, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x66, 0x69, 0x6c, 0x74, 0x65,
	0x72, 0x42, 0x6f, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65,
	0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67,
	0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x8b, 0x01, 0x0a, 0x18, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x74,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x47, 0x0a, 0x0d, 0x62, 0x6f, 0x74, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x0c, 0x62, 0x6f,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65,
	0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x22, 0x56, 0x0a, 0x18, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x42, 0x6f, 0x74, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19,
	0x0a, 0x08, 0x62, 0x6f, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x62, 0x6f, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x22, 0x69, 0x0a, 0x16, 0x53, 0x75,
	0x62, 0x6d, 0x69, 0x74, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x4f, 0x0a, 0x09, 0x68, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e,
	0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x09, 0x68, 0x65, 0x61, 0x72,
	0x74, 0x62, 0x65, 0x61, 0x74, 0x22, 0x19, 0x0a, 0x17, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x48,
	0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x32, 0xbd, 0x03, 0x0a, 0x12, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x62, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x42, 0x6f,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x2c, 0x2e, 0x74, 0x65, 0x6c, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e,
	0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x73, 0x0a, 0x10, 0x4c,
	0x69, 0x73, 0x74, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x12,
	0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x74, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2f, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x74, 0x49,
	0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x5c, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x2f, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x2e, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x42, 0x6f, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x70,
	0x0a, 0x0f, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61,
	0x74, 0x12, 0x2d, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74,
	0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x6d, 0x61, 0x63, 0x68,
	0x69, 0x6e, 0x65, 0x69, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x48,
	0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x56, 0x5a, 0x54, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67,
	0x72, 0x61, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x65, 0x6c,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x69, 0x64, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_teleport_machineid_v1_bot_instance_service_proto_rawDescOnce sync.Once
	file_teleport_machineid_v1_bot_instance_service_proto_rawDescData []byte
)

func file_teleport_machineid_v1_bot_instance_service_proto_rawDescGZIP() []byte {
	file_teleport_machineid_v1_bot_instance_service_proto_rawDescOnce.Do(func() {
		file_teleport_machineid_v1_bot_instance_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_teleport_machineid_v1_bot_instance_service_proto_rawDesc), len(file_teleport_machineid_v1_bot_instance_service_proto_rawDesc)))
	})
	return file_teleport_machineid_v1_bot_instance_service_proto_rawDescData
}

var file_teleport_machineid_v1_bot_instance_service_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_teleport_machineid_v1_bot_instance_service_proto_goTypes = []any{
	(*GetBotInstanceRequest)(nil),      // 0: teleport.machineid.v1.GetBotInstanceRequest
	(*ListBotInstancesRequest)(nil),    // 1: teleport.machineid.v1.ListBotInstancesRequest
	(*ListBotInstancesResponse)(nil),   // 2: teleport.machineid.v1.ListBotInstancesResponse
	(*DeleteBotInstanceRequest)(nil),   // 3: teleport.machineid.v1.DeleteBotInstanceRequest
	(*SubmitHeartbeatRequest)(nil),     // 4: teleport.machineid.v1.SubmitHeartbeatRequest
	(*SubmitHeartbeatResponse)(nil),    // 5: teleport.machineid.v1.SubmitHeartbeatResponse
	(*BotInstance)(nil),                // 6: teleport.machineid.v1.BotInstance
	(*BotInstanceStatusHeartbeat)(nil), // 7: teleport.machineid.v1.BotInstanceStatusHeartbeat
	(*emptypb.Empty)(nil),              // 8: google.protobuf.Empty
}
var file_teleport_machineid_v1_bot_instance_service_proto_depIdxs = []int32{
	6, // 0: teleport.machineid.v1.ListBotInstancesResponse.bot_instances:type_name -> teleport.machineid.v1.BotInstance
	7, // 1: teleport.machineid.v1.SubmitHeartbeatRequest.heartbeat:type_name -> teleport.machineid.v1.BotInstanceStatusHeartbeat
	0, // 2: teleport.machineid.v1.BotInstanceService.GetBotInstance:input_type -> teleport.machineid.v1.GetBotInstanceRequest
	1, // 3: teleport.machineid.v1.BotInstanceService.ListBotInstances:input_type -> teleport.machineid.v1.ListBotInstancesRequest
	3, // 4: teleport.machineid.v1.BotInstanceService.DeleteBotInstance:input_type -> teleport.machineid.v1.DeleteBotInstanceRequest
	4, // 5: teleport.machineid.v1.BotInstanceService.SubmitHeartbeat:input_type -> teleport.machineid.v1.SubmitHeartbeatRequest
	6, // 6: teleport.machineid.v1.BotInstanceService.GetBotInstance:output_type -> teleport.machineid.v1.BotInstance
	2, // 7: teleport.machineid.v1.BotInstanceService.ListBotInstances:output_type -> teleport.machineid.v1.ListBotInstancesResponse
	8, // 8: teleport.machineid.v1.BotInstanceService.DeleteBotInstance:output_type -> google.protobuf.Empty
	5, // 9: teleport.machineid.v1.BotInstanceService.SubmitHeartbeat:output_type -> teleport.machineid.v1.SubmitHeartbeatResponse
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_teleport_machineid_v1_bot_instance_service_proto_init() }
func file_teleport_machineid_v1_bot_instance_service_proto_init() {
	if File_teleport_machineid_v1_bot_instance_service_proto != nil {
		return
	}
	file_teleport_machineid_v1_bot_instance_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_teleport_machineid_v1_bot_instance_service_proto_rawDesc), len(file_teleport_machineid_v1_bot_instance_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_teleport_machineid_v1_bot_instance_service_proto_goTypes,
		DependencyIndexes: file_teleport_machineid_v1_bot_instance_service_proto_depIdxs,
		MessageInfos:      file_teleport_machineid_v1_bot_instance_service_proto_msgTypes,
	}.Build()
	File_teleport_machineid_v1_bot_instance_service_proto = out.File
	file_teleport_machineid_v1_bot_instance_service_proto_goTypes = nil
	file_teleport_machineid_v1_bot_instance_service_proto_depIdxs = nil
}
