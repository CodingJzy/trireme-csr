package client

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
)

// CertificateClient represents a client for the CRD for certs.
type CertificateClient struct {
	restClient rest.Interface
}

// NewClient generates a client for the certificate type.
func NewClient(cfg *rest.Config) (*CertificateClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := certificatev1alpha1.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}

	config := *cfg
	config.GroupVersion = &certificatev1alpha1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	certificateClient := &CertificateClient{
		restClient: client,
	}

	return certificateClient, scheme, nil
}

// Certificates generates an object to communicate with certificates
func (c *CertificateClient) Certificates(namespace string) CertificateInterface {
	return newCertificates(c, namespace)
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *CertificateClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
