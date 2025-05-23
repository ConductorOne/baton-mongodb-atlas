// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: c1/connector/v2/connector.proto

package v2

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Capability int32

const (
	Capability_CAPABILITY_UNSPECIFIED          Capability = 0
	Capability_CAPABILITY_PROVISION            Capability = 1
	Capability_CAPABILITY_SYNC                 Capability = 2
	Capability_CAPABILITY_EVENT_FEED           Capability = 3
	Capability_CAPABILITY_TICKETING            Capability = 4
	Capability_CAPABILITY_ACCOUNT_PROVISIONING Capability = 5
	Capability_CAPABILITY_CREDENTIAL_ROTATION  Capability = 6
	Capability_CAPABILITY_RESOURCE_CREATE      Capability = 7
	Capability_CAPABILITY_RESOURCE_DELETE      Capability = 8
)

// Enum value maps for Capability.
var (
	Capability_name = map[int32]string{
		0: "CAPABILITY_UNSPECIFIED",
		1: "CAPABILITY_PROVISION",
		2: "CAPABILITY_SYNC",
		3: "CAPABILITY_EVENT_FEED",
		4: "CAPABILITY_TICKETING",
		5: "CAPABILITY_ACCOUNT_PROVISIONING",
		6: "CAPABILITY_CREDENTIAL_ROTATION",
		7: "CAPABILITY_RESOURCE_CREATE",
		8: "CAPABILITY_RESOURCE_DELETE",
	}
	Capability_value = map[string]int32{
		"CAPABILITY_UNSPECIFIED":          0,
		"CAPABILITY_PROVISION":            1,
		"CAPABILITY_SYNC":                 2,
		"CAPABILITY_EVENT_FEED":           3,
		"CAPABILITY_TICKETING":            4,
		"CAPABILITY_ACCOUNT_PROVISIONING": 5,
		"CAPABILITY_CREDENTIAL_ROTATION":  6,
		"CAPABILITY_RESOURCE_CREATE":      7,
		"CAPABILITY_RESOURCE_DELETE":      8,
	}
)

func (x Capability) Enum() *Capability {
	p := new(Capability)
	*p = x
	return p
}

func (x Capability) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Capability) Descriptor() protoreflect.EnumDescriptor {
	return file_c1_connector_v2_connector_proto_enumTypes[0].Descriptor()
}

func (Capability) Type() protoreflect.EnumType {
	return &file_c1_connector_v2_connector_proto_enumTypes[0]
}

func (x Capability) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Capability.Descriptor instead.
func (Capability) EnumDescriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{0}
}

type ConnectorMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DisplayName  string                 `protobuf:"bytes,1,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	HelpUrl      string                 `protobuf:"bytes,2,opt,name=help_url,json=helpUrl,proto3" json:"help_url,omitempty"`
	Icon         *AssetRef              `protobuf:"bytes,3,opt,name=icon,proto3" json:"icon,omitempty"`
	Logo         *AssetRef              `protobuf:"bytes,4,opt,name=logo,proto3" json:"logo,omitempty"`
	Profile      *structpb.Struct       `protobuf:"bytes,5,opt,name=profile,proto3" json:"profile,omitempty"`
	Annotations  []*anypb.Any           `protobuf:"bytes,6,rep,name=annotations,proto3" json:"annotations,omitempty"`
	Description  string                 `protobuf:"bytes,7,opt,name=description,proto3" json:"description,omitempty"`
	Capabilities *ConnectorCapabilities `protobuf:"bytes,8,opt,name=capabilities,proto3" json:"capabilities,omitempty"`
}

func (x *ConnectorMetadata) Reset() {
	*x = ConnectorMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectorMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectorMetadata) ProtoMessage() {}

func (x *ConnectorMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectorMetadata.ProtoReflect.Descriptor instead.
func (*ConnectorMetadata) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{0}
}

func (x *ConnectorMetadata) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *ConnectorMetadata) GetHelpUrl() string {
	if x != nil {
		return x.HelpUrl
	}
	return ""
}

func (x *ConnectorMetadata) GetIcon() *AssetRef {
	if x != nil {
		return x.Icon
	}
	return nil
}

func (x *ConnectorMetadata) GetLogo() *AssetRef {
	if x != nil {
		return x.Logo
	}
	return nil
}

func (x *ConnectorMetadata) GetProfile() *structpb.Struct {
	if x != nil {
		return x.Profile
	}
	return nil
}

func (x *ConnectorMetadata) GetAnnotations() []*anypb.Any {
	if x != nil {
		return x.Annotations
	}
	return nil
}

func (x *ConnectorMetadata) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *ConnectorMetadata) GetCapabilities() *ConnectorCapabilities {
	if x != nil {
		return x.Capabilities
	}
	return nil
}

type ConnectorCapabilities struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResourceTypeCapabilities []*ResourceTypeCapability `protobuf:"bytes,1,rep,name=resource_type_capabilities,json=resourceTypeCapabilities,proto3" json:"resource_type_capabilities,omitempty"`
	ConnectorCapabilities    []Capability              `protobuf:"varint,2,rep,packed,name=connector_capabilities,json=connectorCapabilities,proto3,enum=c1.connector.v2.Capability" json:"connector_capabilities,omitempty"`
}

func (x *ConnectorCapabilities) Reset() {
	*x = ConnectorCapabilities{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectorCapabilities) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectorCapabilities) ProtoMessage() {}

func (x *ConnectorCapabilities) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectorCapabilities.ProtoReflect.Descriptor instead.
func (*ConnectorCapabilities) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{1}
}

func (x *ConnectorCapabilities) GetResourceTypeCapabilities() []*ResourceTypeCapability {
	if x != nil {
		return x.ResourceTypeCapabilities
	}
	return nil
}

func (x *ConnectorCapabilities) GetConnectorCapabilities() []Capability {
	if x != nil {
		return x.ConnectorCapabilities
	}
	return nil
}

type ResourceTypeCapability struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResourceType *ResourceType `protobuf:"bytes,1,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
	Capabilities []Capability  `protobuf:"varint,2,rep,packed,name=capabilities,proto3,enum=c1.connector.v2.Capability" json:"capabilities,omitempty"`
}

func (x *ResourceTypeCapability) Reset() {
	*x = ResourceTypeCapability{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceTypeCapability) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceTypeCapability) ProtoMessage() {}

func (x *ResourceTypeCapability) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceTypeCapability.ProtoReflect.Descriptor instead.
func (*ResourceTypeCapability) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{2}
}

func (x *ResourceTypeCapability) GetResourceType() *ResourceType {
	if x != nil {
		return x.ResourceType
	}
	return nil
}

func (x *ResourceTypeCapability) GetCapabilities() []Capability {
	if x != nil {
		return x.Capabilities
	}
	return nil
}

type ConnectorServiceGetMetadataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ConnectorServiceGetMetadataRequest) Reset() {
	*x = ConnectorServiceGetMetadataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectorServiceGetMetadataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectorServiceGetMetadataRequest) ProtoMessage() {}

func (x *ConnectorServiceGetMetadataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectorServiceGetMetadataRequest.ProtoReflect.Descriptor instead.
func (*ConnectorServiceGetMetadataRequest) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{3}
}

type ConnectorServiceGetMetadataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata *ConnectorMetadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *ConnectorServiceGetMetadataResponse) Reset() {
	*x = ConnectorServiceGetMetadataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectorServiceGetMetadataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectorServiceGetMetadataResponse) ProtoMessage() {}

func (x *ConnectorServiceGetMetadataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectorServiceGetMetadataResponse.ProtoReflect.Descriptor instead.
func (*ConnectorServiceGetMetadataResponse) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{4}
}

func (x *ConnectorServiceGetMetadataResponse) GetMetadata() *ConnectorMetadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type ConnectorServiceValidateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ConnectorServiceValidateRequest) Reset() {
	*x = ConnectorServiceValidateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectorServiceValidateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectorServiceValidateRequest) ProtoMessage() {}

func (x *ConnectorServiceValidateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectorServiceValidateRequest.ProtoReflect.Descriptor instead.
func (*ConnectorServiceValidateRequest) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{5}
}

// NOTE(morgabra) We're expecting correct grpc.Status responses
// for things like 401/403/500, etc
type ConnectorServiceValidateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Annotations []*anypb.Any `protobuf:"bytes,1,rep,name=annotations,proto3" json:"annotations,omitempty"`
}

func (x *ConnectorServiceValidateResponse) Reset() {
	*x = ConnectorServiceValidateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_c1_connector_v2_connector_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectorServiceValidateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectorServiceValidateResponse) ProtoMessage() {}

func (x *ConnectorServiceValidateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_c1_connector_v2_connector_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectorServiceValidateResponse.ProtoReflect.Descriptor instead.
func (*ConnectorServiceValidateResponse) Descriptor() ([]byte, []int) {
	return file_c1_connector_v2_connector_proto_rawDescGZIP(), []int{6}
}

func (x *ConnectorServiceValidateResponse) GetAnnotations() []*anypb.Any {
	if x != nil {
		return x.Annotations
	}
	return nil
}

var File_c1_connector_v2_connector_proto protoreflect.FileDescriptor

var file_c1_connector_v2_connector_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x76,
	0x32, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0f, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e,
	0x76, 0x32, 0x1a, 0x1b, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2f, 0x76, 0x32, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1e, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x76, 0x32,
	0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xbf, 0x03, 0x0a, 0x11, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x2d, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c,
	0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0a, 0xfa,
	0x42, 0x07, 0x72, 0x05, 0x20, 0x01, 0x28, 0x80, 0x08, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70, 0x6c,
	0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x68, 0x65, 0x6c, 0x70, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1a, 0xfa, 0x42, 0x17, 0x72, 0x15, 0x20,
	0x01, 0x28, 0x80, 0x08, 0x3a, 0x08, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0xd0, 0x01,
	0x01, 0x88, 0x01, 0x01, 0x52, 0x07, 0x68, 0x65, 0x6c, 0x70, 0x55, 0x72, 0x6c, 0x12, 0x2d, 0x0a,
	0x04, 0x69, 0x63, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x41, 0x73,
	0x73, 0x65, 0x74, 0x52, 0x65, 0x66, 0x52, 0x04, 0x69, 0x63, 0x6f, 0x6e, 0x12, 0x2d, 0x0a, 0x04,
	0x6c, 0x6f, 0x67, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31, 0x2e,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x41, 0x73, 0x73,
	0x65, 0x74, 0x52, 0x65, 0x66, 0x52, 0x04, 0x6c, 0x6f, 0x67, 0x6f, 0x12, 0x31, 0x0a, 0x07, 0x70,
	0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x36,
	0x0a, 0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x06, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x2f, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0d, 0xfa, 0x42, 0x0a,
	0x72, 0x08, 0x20, 0x01, 0x28, 0x80, 0x20, 0xd0, 0x01, 0x01, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4a, 0x0a, 0x0c, 0x63, 0x61, 0x70, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e,
	0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x0c, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x69, 0x65, 0x73, 0x22, 0xd2, 0x01, 0x0a, 0x15, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x65, 0x0a,
	0x1a, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x63,
	0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x27, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x52, 0x18, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x69, 0x65, 0x73, 0x12, 0x52, 0x0a, 0x16, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x5f, 0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x52, 0x15, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x43, 0x61, 0x70, 0x61,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x22, 0x9d, 0x01, 0x0a, 0x16, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x12, 0x42, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x31, 0x2e,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3f, 0x0a, 0x0c, 0x63, 0x61, 0x70, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x1b, 0x2e,
	0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e,
	0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x52, 0x0c, 0x63, 0x61, 0x70, 0x61,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x22, 0x24, 0x0a, 0x22, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x47, 0x65, 0x74, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x65,
	0x0a, 0x23, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3e, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x21, 0x0a, 0x1f, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x5a, 0x0a, 0x20, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x36, 0x0a, 0x0b,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x0b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2a, 0x95, 0x02, 0x0a, 0x0a, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x12, 0x1a, 0x0a, 0x16, 0x43, 0x41, 0x50, 0x41, 0x42, 0x49, 0x4c, 0x49, 0x54,
	0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x18, 0x0a, 0x14, 0x43, 0x41, 0x50, 0x41, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x50, 0x52,
	0x4f, 0x56, 0x49, 0x53, 0x49, 0x4f, 0x4e, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x43, 0x41, 0x50,
	0x41, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x53, 0x59, 0x4e, 0x43, 0x10, 0x02, 0x12, 0x19,
	0x0a, 0x15, 0x43, 0x41, 0x50, 0x41, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x45, 0x56, 0x45,
	0x4e, 0x54, 0x5f, 0x46, 0x45, 0x45, 0x44, 0x10, 0x03, 0x12, 0x18, 0x0a, 0x14, 0x43, 0x41, 0x50,
	0x41, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x54, 0x49, 0x43, 0x4b, 0x45, 0x54, 0x49, 0x4e,
	0x47, 0x10, 0x04, 0x12, 0x23, 0x0a, 0x1f, 0x43, 0x41, 0x50, 0x41, 0x42, 0x49, 0x4c, 0x49, 0x54,
	0x59, 0x5f, 0x41, 0x43, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x5f, 0x50, 0x52, 0x4f, 0x56, 0x49, 0x53,
	0x49, 0x4f, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x05, 0x12, 0x22, 0x0a, 0x1e, 0x43, 0x41, 0x50, 0x41,
	0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x43, 0x52, 0x45, 0x44, 0x45, 0x4e, 0x54, 0x49, 0x41,
	0x4c, 0x5f, 0x52, 0x4f, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x06, 0x12, 0x1e, 0x0a, 0x1a,
	0x43, 0x41, 0x50, 0x41, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x52, 0x45, 0x53, 0x4f, 0x55,
	0x52, 0x43, 0x45, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x10, 0x07, 0x12, 0x1e, 0x0a, 0x1a,
	0x43, 0x41, 0x50, 0x41, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x52, 0x45, 0x53, 0x4f, 0x55,
	0x52, 0x43, 0x45, 0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x08, 0x32, 0xfd, 0x01, 0x0a,
	0x10, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x78, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x33, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e,
	0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x34, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6f, 0x0a, 0x08, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x12, 0x30, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e, 0x63, 0x31, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x36, 0x5a, 0x34,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x64, 0x75,
	0x63, 0x74, 0x6f, 0x72, 0x6f, 0x6e, 0x65, 0x2f, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2d, 0x73, 0x64,
	0x6b, 0x2f, 0x70, 0x62, 0x2f, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x2f, 0x76, 0x32, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_c1_connector_v2_connector_proto_rawDescOnce sync.Once
	file_c1_connector_v2_connector_proto_rawDescData = file_c1_connector_v2_connector_proto_rawDesc
)

func file_c1_connector_v2_connector_proto_rawDescGZIP() []byte {
	file_c1_connector_v2_connector_proto_rawDescOnce.Do(func() {
		file_c1_connector_v2_connector_proto_rawDescData = protoimpl.X.CompressGZIP(file_c1_connector_v2_connector_proto_rawDescData)
	})
	return file_c1_connector_v2_connector_proto_rawDescData
}

var file_c1_connector_v2_connector_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_c1_connector_v2_connector_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_c1_connector_v2_connector_proto_goTypes = []interface{}{
	(Capability)(0),                             // 0: c1.connector.v2.Capability
	(*ConnectorMetadata)(nil),                   // 1: c1.connector.v2.ConnectorMetadata
	(*ConnectorCapabilities)(nil),               // 2: c1.connector.v2.ConnectorCapabilities
	(*ResourceTypeCapability)(nil),              // 3: c1.connector.v2.ResourceTypeCapability
	(*ConnectorServiceGetMetadataRequest)(nil),  // 4: c1.connector.v2.ConnectorServiceGetMetadataRequest
	(*ConnectorServiceGetMetadataResponse)(nil), // 5: c1.connector.v2.ConnectorServiceGetMetadataResponse
	(*ConnectorServiceValidateRequest)(nil),     // 6: c1.connector.v2.ConnectorServiceValidateRequest
	(*ConnectorServiceValidateResponse)(nil),    // 7: c1.connector.v2.ConnectorServiceValidateResponse
	(*AssetRef)(nil),                            // 8: c1.connector.v2.AssetRef
	(*structpb.Struct)(nil),                     // 9: google.protobuf.Struct
	(*anypb.Any)(nil),                           // 10: google.protobuf.Any
	(*ResourceType)(nil),                        // 11: c1.connector.v2.ResourceType
}
var file_c1_connector_v2_connector_proto_depIdxs = []int32{
	8,  // 0: c1.connector.v2.ConnectorMetadata.icon:type_name -> c1.connector.v2.AssetRef
	8,  // 1: c1.connector.v2.ConnectorMetadata.logo:type_name -> c1.connector.v2.AssetRef
	9,  // 2: c1.connector.v2.ConnectorMetadata.profile:type_name -> google.protobuf.Struct
	10, // 3: c1.connector.v2.ConnectorMetadata.annotations:type_name -> google.protobuf.Any
	2,  // 4: c1.connector.v2.ConnectorMetadata.capabilities:type_name -> c1.connector.v2.ConnectorCapabilities
	3,  // 5: c1.connector.v2.ConnectorCapabilities.resource_type_capabilities:type_name -> c1.connector.v2.ResourceTypeCapability
	0,  // 6: c1.connector.v2.ConnectorCapabilities.connector_capabilities:type_name -> c1.connector.v2.Capability
	11, // 7: c1.connector.v2.ResourceTypeCapability.resource_type:type_name -> c1.connector.v2.ResourceType
	0,  // 8: c1.connector.v2.ResourceTypeCapability.capabilities:type_name -> c1.connector.v2.Capability
	1,  // 9: c1.connector.v2.ConnectorServiceGetMetadataResponse.metadata:type_name -> c1.connector.v2.ConnectorMetadata
	10, // 10: c1.connector.v2.ConnectorServiceValidateResponse.annotations:type_name -> google.protobuf.Any
	4,  // 11: c1.connector.v2.ConnectorService.GetMetadata:input_type -> c1.connector.v2.ConnectorServiceGetMetadataRequest
	6,  // 12: c1.connector.v2.ConnectorService.Validate:input_type -> c1.connector.v2.ConnectorServiceValidateRequest
	5,  // 13: c1.connector.v2.ConnectorService.GetMetadata:output_type -> c1.connector.v2.ConnectorServiceGetMetadataResponse
	7,  // 14: c1.connector.v2.ConnectorService.Validate:output_type -> c1.connector.v2.ConnectorServiceValidateResponse
	13, // [13:15] is the sub-list for method output_type
	11, // [11:13] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_c1_connector_v2_connector_proto_init() }
func file_c1_connector_v2_connector_proto_init() {
	if File_c1_connector_v2_connector_proto != nil {
		return
	}
	file_c1_connector_v2_asset_proto_init()
	file_c1_connector_v2_resource_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_c1_connector_v2_connector_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectorMetadata); i {
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
		file_c1_connector_v2_connector_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectorCapabilities); i {
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
		file_c1_connector_v2_connector_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceTypeCapability); i {
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
		file_c1_connector_v2_connector_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectorServiceGetMetadataRequest); i {
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
		file_c1_connector_v2_connector_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectorServiceGetMetadataResponse); i {
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
		file_c1_connector_v2_connector_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectorServiceValidateRequest); i {
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
		file_c1_connector_v2_connector_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectorServiceValidateResponse); i {
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
			RawDescriptor: file_c1_connector_v2_connector_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_c1_connector_v2_connector_proto_goTypes,
		DependencyIndexes: file_c1_connector_v2_connector_proto_depIdxs,
		EnumInfos:         file_c1_connector_v2_connector_proto_enumTypes,
		MessageInfos:      file_c1_connector_v2_connector_proto_msgTypes,
	}.Build()
	File_c1_connector_v2_connector_proto = out.File
	file_c1_connector_v2_connector_proto_rawDesc = nil
	file_c1_connector_v2_connector_proto_goTypes = nil
	file_c1_connector_v2_connector_proto_depIdxs = nil
}
