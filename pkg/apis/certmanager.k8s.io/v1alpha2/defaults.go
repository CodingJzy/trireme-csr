package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

// SetDefaults_Certificate ensures defaults are set
func SetDefaults_Certificate(obj *Certificate) {
	if obj.Status.Phase == "" {
		obj.Status.Phase = CertificateUnknown
		if obj.Status.Reason == "" {
			obj.Status.Reason = StatusReasonUnprocessed
		}
		if obj.Status.Message == "" {
			obj.Status.Message = "No Status Phase was defined"
		}
	}
}
