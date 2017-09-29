package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CertificateResourcePlural = "certificates"

type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CertificateSpec   `json:"spec"`
	Status            CertificateStatus `json:"status,omitempty"`
}

// CertificateSpec is the specification for Certificates on the API
type CertificateSpec struct {
	Foo string `json:"foo"`
	Bar bool   `json:"bar"`
}

// CertificateStatus is the status for Certificates on the API
type CertificateStatus struct {
	State       CertificateState `json:"state,omitempty"`
	Certificate string           `json:"message,omitempty"`
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
	metav1.ListMeta `json:"metadata"`
	Items           []Certificate `json:"items"`
}
