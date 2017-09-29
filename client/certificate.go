package client

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	certificatev1 "github.com/aporeto-inc/trireme-csr/apis/v1"
)

// CertificateInterface is the interface to the certificates
type CertificateInterface interface {
	Create(*certificatev1.Certificate) (*certificatev1.Certificate, error)
	Update(*certificatev1.Certificate) (*certificatev1.Certificate, error)
	//UpdateStatus(*certificatev1.Certificate) (*certificatev1.Certificate, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	//DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*certificatev1.Certificate, error)
	List(opts meta_v1.ListOptions) (*certificatev1.CertificateList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	//Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *certificatev1.Certificate, err error)
}

// certificates implements certificateInterface
type certificates struct {
	client rest.Interface
	ns     string
}

// newcertificates returns a certificates
func newCertificates(c *CertificateClient, namespace string) *certificates {
	return &certificates{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *certificates) Create(certificate *certificatev1.Certificate) (result *certificatev1.Certificate, err error) {
	result = &certificatev1.Certificate{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("certificates").
		Body(certificate).
		Do().
		Into(result)
	return
}

// Update takes the representation of a certificate and updates it. Returns the server's representation of the certificate, and an error, if there is any.
func (c *certificates) Update(certificate *certificatev1.Certificate) (result *certificatev1.Certificate, err error) {
	result = &certificatev1.Certificate{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("certificates").
		Name(certificate.Name).
		Body(certificate).
		Do().
		Into(result)
	return
}

// Delete takes name of the certificate and deletes it. Returns an error if one occurs.
func (c *certificates) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("certificates").
		Name(name).
		Body(options).
		Do().
		Error()
}

// Get takes name of the certificate, and returns the corresponding certificate object, and an error if there is any.
func (c *certificates) Get(name string, options meta_v1.GetOptions) (result *certificatev1.Certificate, err error) {
	result = &certificatev1.Certificate{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("certificates").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of certificates that match those selectors.
func (c *certificates) List(opts meta_v1.ListOptions) (result *certificatev1.CertificateList, err error) {
	result = &certificatev1.CertificateList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("certificates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested certificates.
func (c *certificates) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("certificates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}
