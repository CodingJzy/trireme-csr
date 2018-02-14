package client

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	certificatev1alpha2 "github.com/aporeto-inc/trireme-csr/apis/v1alpha2"
)

// CertificateInterface is the interface to the certificates
type CertificateInterface interface {
	Create(*certificatev1alpha2.Certificate) (*certificatev1alpha2.Certificate, error)
	Update(*certificatev1alpha2.Certificate) (*certificatev1alpha2.Certificate, error)
	UpdateStatus(*certificatev1alpha2.Certificate) (*certificatev1alpha2.Certificate, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	//DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*certificatev1alpha2.Certificate, error)
	List(opts meta_v1.ListOptions) (*certificatev1alpha2.CertificateList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	//Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *certificatev1alpha2.Certificate, err error)
}

// certificates implements certificateInterface
type certificates struct {
	client rest.Interface
}

// newcertificates returns a certificates
func newCertificates(c *CertificateClient) *certificates {
	return &certificates{
		client: c.RESTClient(),
	}
}

func (c *certificates) Create(certificate *certificatev1alpha2.Certificate) (result *certificatev1alpha2.Certificate, err error) {
	result = &certificatev1alpha2.Certificate{}
	err = c.client.Post().
		Resource("certificates").
		Body(certificate).
		Do().
		Into(result)
	return
}

// Update takes the representation of a certificate and updates it. Returns the server's representation of the certificate, and an error, if there is any.
func (c *certificates) Update(certificate *certificatev1alpha2.Certificate) (result *certificatev1alpha2.Certificate, err error) {
	result = &certificatev1alpha2.Certificate{}
	err = c.client.Put().
		Resource("certificates").
		Name(certificate.Name).
		Body(certificate).
		Do().
		Into(result)
	return
}

// Update takes the representation of a certificate and updates it. Returns the server's representation of the certificate, and an error, if there is any.
func (c *certificates) UpdateStatus(certificate *certificatev1alpha2.Certificate) (result *certificatev1alpha2.Certificate, err error) {
	result = &certificatev1alpha2.Certificate{}
	err = c.client.Put().
		Resource("certificates").
		Name(certificate.Name).
		SubResource("status").
		Body(certificate).
		Do().
		Into(result)
	return
}

// Delete takes name of the certificate and deletes it. Returns an error if one occurs.
func (c *certificates) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("certificates").
		Name(name).
		Do().
		Error()
}

// Get takes name of the certificate, and returns the corresponding certificate object, and an error if there is any.
func (c *certificates) Get(name string, options meta_v1.GetOptions) (result *certificatev1alpha2.Certificate, err error) {
	result = &certificatev1alpha2.Certificate{}
	err = c.client.Get().
		Resource("certificates").
		Name(name).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of certificates that match those selectors.
func (c *certificates) List(opts meta_v1.ListOptions) (result *certificatev1alpha2.CertificateList, err error) {
	result = &certificatev1alpha2.CertificateList{}
	err = c.client.Get().
		Resource("certificates").
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested certificates.
func (c *certificates) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("certificates").
		Watch()
}
