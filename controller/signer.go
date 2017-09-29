package signer

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type signerController struct {
	signerClient *rest.RESTClient
}

func (c *signerController) Run() error {}

func (c *signerController) watchCerts() (cache.Controller, error) {
	source := cache.NewListWatchFromClient(
		c.ExampleClient,
		crv1.ExampleResourcePlural,
		apiv1.NamespaceAll,
		fields.Everything())

	_, controller := cache.NewInformer(
		source,

		// The object type.
		&crv1.Example{},

		0,

		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	go controller.Run(ctx.Done())
	return controller, nil
}

func (c *ExampleController) onAdd(obj interface{})

func (c *ExampleController) onUpdate(oldObj, newObj interface{}) {}

func (c *ExampleController) onDelete(obj interface{}) {}
