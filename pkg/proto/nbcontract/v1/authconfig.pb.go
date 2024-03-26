// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: pkg/proto/nbcontract/v1/authconfig.proto

package nbcontractv1

import (
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

// Auth Config fields stored in azure.json used by cloud-provider-azure
type AuthConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TargetCloud                 string `protobuf:"bytes,1,opt,name=target_cloud,json=targetCloud,proto3" json:"target_cloud,omitempty"` // set to cloud, can probably get rid of this, analyze more
	TenantId                    string `protobuf:"bytes,2,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	SubscriptionId              string `protobuf:"bytes,3,opt,name=subscription_id,json=subscriptionId,proto3" json:"subscription_id,omitempty"`
	ServicePrincipalId          string `protobuf:"bytes,4,opt,name=service_principal_id,json=servicePrincipalId,proto3" json:"service_principal_id,omitempty"`             // set to aadClientId
	ServicePrincipalSecret      string `protobuf:"bytes,5,opt,name=service_principal_secret,json=servicePrincipalSecret,proto3" json:"service_principal_secret,omitempty"` // set to aadClientSecret
	AssignedIdentityId          string `protobuf:"bytes,6,opt,name=assigned_identity_id,json=assignedIdentityId,proto3" json:"assigned_identity_id,omitempty"`             //could be user or system assigned, depending on the type
	UseManagedIdentityExtension bool   `protobuf:"varint,7,opt,name=use_managed_identity_extension,json=useManagedIdentityExtension,proto3" json:"use_managed_identity_extension,omitempty"`
}

func (x *AuthConfig) Reset() {
	*x = AuthConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_nbcontract_v1_authconfig_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthConfig) ProtoMessage() {}

func (x *AuthConfig) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_nbcontract_v1_authconfig_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthConfig.ProtoReflect.Descriptor instead.
func (*AuthConfig) Descriptor() ([]byte, []int) {
	return file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescGZIP(), []int{0}
}

func (x *AuthConfig) GetTargetCloud() string {
	if x != nil {
		return x.TargetCloud
	}
	return ""
}

func (x *AuthConfig) GetTenantId() string {
	if x != nil {
		return x.TenantId
	}
	return ""
}

func (x *AuthConfig) GetSubscriptionId() string {
	if x != nil {
		return x.SubscriptionId
	}
	return ""
}

func (x *AuthConfig) GetServicePrincipalId() string {
	if x != nil {
		return x.ServicePrincipalId
	}
	return ""
}

func (x *AuthConfig) GetServicePrincipalSecret() string {
	if x != nil {
		return x.ServicePrincipalSecret
	}
	return ""
}

func (x *AuthConfig) GetAssignedIdentityId() string {
	if x != nil {
		return x.AssignedIdentityId
	}
	return ""
}

func (x *AuthConfig) GetUseManagedIdentityExtension() bool {
	if x != nil {
		return x.UseManagedIdentityExtension
	}
	return false
}

var File_pkg_proto_nbcontract_v1_authconfig_proto protoreflect.FileDescriptor

var file_pkg_proto_nbcontract_v1_authconfig_proto_rawDesc = []byte{
	0x0a, 0x28, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x62, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x6e, 0x62, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2e, 0x76, 0x31, 0x22, 0xd8, 0x02, 0x0a, 0x0a, 0x41, 0x75,
	0x74, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x21, 0x0a, 0x0c, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x5f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x74,
	0x65, 0x6e, 0x61, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x75, 0x62, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0e, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x12, 0x30, 0x0a, 0x14, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x70, 0x72, 0x69,
	0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x12, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61,
	0x6c, 0x49, 0x64, 0x12, 0x38, 0x0a, 0x18, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x70,
	0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x5f, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72,
	0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x12, 0x30, 0x0a,
	0x14, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x61, 0x73, 0x73,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x49, 0x64, 0x12,
	0x43, 0x0a, 0x1e, 0x75, 0x73, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x64, 0x5f, 0x69,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1b, 0x75, 0x73, 0x65, 0x4d, 0x61, 0x6e, 0x61,
	0x67, 0x65, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x45, 0x78, 0x74, 0x65, 0x6e,
	0x73, 0x69, 0x6f, 0x6e, 0x42, 0xbb, 0x01, 0x0a, 0x11, 0x63, 0x6f, 0x6d, 0x2e, 0x6e, 0x62, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2e, 0x76, 0x31, 0x42, 0x0f, 0x41, 0x75, 0x74, 0x68,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x40, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x41, 0x7a, 0x75, 0x72, 0x65, 0x2f,
	0x41, 0x67, 0x65, 0x6e, 0x74, 0x42, 0x61, 0x6b, 0x65, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f,
	0x76, 0x31, 0x3b, 0x6e, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x76, 0x31, 0xa2,
	0x02, 0x03, 0x4e, 0x58, 0x58, 0xaa, 0x02, 0x0d, 0x4e, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0d, 0x4e, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x19, 0x4e, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x0e, 0x4e, 0x62, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x3a, 0x3a,
	0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescOnce sync.Once
	file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescData = file_pkg_proto_nbcontract_v1_authconfig_proto_rawDesc
)

func file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescGZIP() []byte {
	file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescOnce.Do(func() {
		file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescData)
	})
	return file_pkg_proto_nbcontract_v1_authconfig_proto_rawDescData
}

var file_pkg_proto_nbcontract_v1_authconfig_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_pkg_proto_nbcontract_v1_authconfig_proto_goTypes = []interface{}{
	(*AuthConfig)(nil), // 0: nbcontract.v1.AuthConfig
}
var file_pkg_proto_nbcontract_v1_authconfig_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_proto_nbcontract_v1_authconfig_proto_init() }
func file_pkg_proto_nbcontract_v1_authconfig_proto_init() {
	if File_pkg_proto_nbcontract_v1_authconfig_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_proto_nbcontract_v1_authconfig_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthConfig); i {
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
			RawDescriptor: file_pkg_proto_nbcontract_v1_authconfig_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_proto_nbcontract_v1_authconfig_proto_goTypes,
		DependencyIndexes: file_pkg_proto_nbcontract_v1_authconfig_proto_depIdxs,
		MessageInfos:      file_pkg_proto_nbcontract_v1_authconfig_proto_msgTypes,
	}.Build()
	File_pkg_proto_nbcontract_v1_authconfig_proto = out.File
	file_pkg_proto_nbcontract_v1_authconfig_proto_rawDesc = nil
	file_pkg_proto_nbcontract_v1_authconfig_proto_goTypes = nil
	file_pkg_proto_nbcontract_v1_authconfig_proto_depIdxs = nil
}
