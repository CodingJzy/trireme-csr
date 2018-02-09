package v1alpha2

import (
	"crypto/x509"
	"fmt"

	"github.com/aporeto-inc/tg/tglib"
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
	Message     string           `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	Reason      string           `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
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
	StatusReasonSubmitted                     = "Submitted"
	StatusReasonProcessedApprovedSignedIssued = "ProcessedApprovedSignedIssued"
	StatusReasonProcessedRejected             = "ProcessedRejected"
	StatusReasonProcessedRejectedInvalidCSR   = "ProcessedRejectedInvalidCSR"
	StatusReasonProcessedRejectedInvalidCerts = "ProcessedRejectedInvalidCerts"
)

// CertificateList represents a list of certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Certificate `json:"items" protobuf:"bytes,2,rep,name=items"`
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

// GetCertificateRequest returns a `*x509.CertificateRequest` object from the spec, or
// an error if this fails.
func (c *CertificateSpec) GetCertificateRequest() (*x509.CertificateRequest, error) {
	if c.Request == nil {
		return nil, fmt.Errorf("no certificate request in spec")
	}
	csrs, err := tglib.LoadCSRs(c.Request)
	if err != nil {
		return nil, err
	}
	if len(csrs) != 1 {
		return nil, fmt.Errorf("spec must contain exactly one CSR")
	}
	return csrs[0], nil
}

// GetCertificate returns a `*x509.Certificate` object from the status holding the
// issued certificate from the CA, or an error if this fails
func (c *CertificateStatus) GetCertificate() (*x509.Certificate, error) {
	if c.Certificate == nil {
		return nil, fmt.Errorf("no certificate has been issued yet")
	}
	return tglib.ReadCertificatePEMFromData(c.Certificate)
}

// GetCACertificate returns a `*x509.Certificate` object from the status holding the issuing CA
// certificate, or an error if this fails
func (c *CertificateStatus) GetCACertificate() (*x509.Certificate, error) {
	if c.Ca == nil {
		return nil, fmt.Errorf("no CA certificate found")
	}
	return tglib.ReadCertificatePEMFromData(c.Ca)
}

// GetCertificateRequest returns a `*x509.CertificateRequest` object from the spec, or
// an error if this fails.
func (c *Certificate) GetCertificateRequest() (*x509.CertificateRequest, error) {
	return c.Spec.GetCertificateRequest()
}

// GetCertificate returns a `*x509.Certificate` object from the status holding the
// issued certificate from the CA, or an error if this fails
func (c *Certificate) GetCertificate() (*x509.Certificate, error) {
	return c.Status.GetCertificate()
}

// GetCACertificate returns a `*x509.Certificate` object from the status holding the issuing CA
// certificate, or an error if this fails
func (c *Certificate) GetCACertificate() (*x509.Certificate, error) {
	return c.Status.GetCACertificate()
}
