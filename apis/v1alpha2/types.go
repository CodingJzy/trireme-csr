package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CertificateResourcePlural is the ressource name used to get a list of cetts.
const CertificateResourcePlural = "certificates"

// Certificate is the specification for the Certificate object on the Kubernetes API
type Certificate struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata"`
	Spec              CertificateSpec `json:"spec"`
	// +optional
	Status CertificateStatus `json:"status,omitempty"`
}

// CertificateSpec is the specification for Certificates on the API
type CertificateSpec struct {
	// Base64-encoded PKCS#10 CSR data
	Request []byte `json:"request" protobuf:"bytes,1,opt,name=request"`
}

// CertificateStatus is the status for Certificates on the API
type CertificateStatus struct {
	State       CertificateState `json:"state,omitempty"`
	Certificate []byte           `json:"certificate,omitempty" protobuf:"bytes,2,opt,name=certificate"`
	Token       []byte           `json:"token,omitempty" protobuf:"bytes,2,opt,name=token"`
	Ca          []byte           `json:"ca,omitempty" protobuf:"bytes,2,opt,name=ca"`
}

// CertificateState defines the state of the certificate
type CertificateState string

const (
	// CertificateStateCreated defines that the Certificate was created
	CertificateStateCreated CertificateState = "Created"
	// CertificateStateProcessed defines that the certificate was processed
	CertificateStateProcessed CertificateState = "Processed"
)

// CertificateList represents a list of certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata"`
	Items           []Certificate `json:"items"`
}

// DeepCopyObject returns a copy of the object
func (c *Certificate) DeepCopyObject() runtime.Object {
	// TODO: Correct DeepCopy
	return &Certificate{}
}

// DeepCopyObject returns a copy of the object
func (c *CertificateList) DeepCopyObject() runtime.Object {
	// TODO: Correct DeepCopy
	return &CertificateList{}
}

// DeepCopyObject returns a copy of the object
func (c *CertificateStatus) DeepCopyObject() *CertificateStatus {
	// TODO: Correct DeepCopy
	return &CertificateStatus{}
}

// DeepCopyObject returns a copy of the object
func (c CertificateState) DeepCopyObject() CertificateState {
	// TODO: Correct DeepCopy
	return c
}

// DeepCopyObject returns a copy of the object
func (c *CertificateSpec) DeepCopyObject() *CertificateSpec {
	// TODO: Correct DeepCopy
	return &CertificateSpec{}
}
