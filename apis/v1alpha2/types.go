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
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              CertificateSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// +optional
	Status CertificateStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// CertificateSpec is the specification for Certificates on the API
type CertificateSpec struct {
	// Base64-encoded PKCS#10 CSR data
	Request []byte `json:"request" protobuf:"bytes,1,opt,name=request"`
}

// CertificateStatus is the status for Certificates on the API
type CertificateStatus struct {
	Phase       CertificatePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=CertificatePhase"`
	Message     string           `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	Reason      string           `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	Certificate []byte           `json:"certificate,omitempty" protobuf:"bytes,4,opt,name=certificate"`
	Token       []byte           `json:"token,omitempty" protobuf:"bytes,5,opt,name=token"`
	Ca          []byte           `json:"ca,omitempty" protobuf:"bytes,6,opt,name=ca"`
}

// CertificatePhase defines the phase of the certificate
type CertificatePhase string

const (
	// CertificateSubmitted defines that the CSR was submitted
	CertificateSubmitted CertificatePhase = "Submitted"
	// CertificateSigned defines that the CSR was processed, and the request has been approved and the certificate was signed and has been issued
	CertificateSigned CertificatePhase = "Signed"
	// CertificateRejected defines that the CSR was processed, and the request has been rejected and therefore no certificate was issued
	CertificateRejected CertificatePhase = "Rejected"
	// CertificateUnknown defines that the CSR is in an unknown state, and the controller will not take any further action on this object
	CertificateUnknown CertificatePhase = "Unknown"
)

// Certificate Status reasons
const (
	StatusReasonUnprocessed                   = "Unprocessed"
	StatusReasonProcessedApprovedSignedIssued = "ProcessedApprovedSignedIssued"
	StatusReasonProcessedRejected             = "ProcessedRejected"
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
func (c CertificatePhase) DeepCopyObject() CertificatePhase {
	// TODO: Correct DeepCopy
	return c
}

// DeepCopyObject returns a copy of the object
func (c *CertificateSpec) DeepCopyObject() *CertificateSpec {
	// TODO: Correct DeepCopy
	return &CertificateSpec{}
}
