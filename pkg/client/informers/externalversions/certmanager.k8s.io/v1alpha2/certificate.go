//
// Copyright 2017 Aporeto, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// This file was automatically generated by informer-gen

package v1alpha2

import (
	time "time"

	certmanager_k8s_io_v1alpha2 "github.com/aporeto-inc/trireme-csr/pkg/apis/certmanager.k8s.io/v1alpha2"
	versioned "github.com/aporeto-inc/trireme-csr/pkg/client/clientset/versioned"
	internalinterfaces "github.com/aporeto-inc/trireme-csr/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha2 "github.com/aporeto-inc/trireme-csr/pkg/client/listers/certmanager.k8s.io/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CertificateInformer provides access to a shared informer and lister for
// Certificates.
type CertificateInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.CertificateLister
}

type certificateInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewCertificateInformer constructs a new informer for Certificate type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCertificateInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCertificateInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredCertificateInformer constructs a new informer for Certificate type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCertificateInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CertmanagerV1alpha2().Certificates().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CertmanagerV1alpha2().Certificates().Watch(options)
			},
		},
		&certmanager_k8s_io_v1alpha2.Certificate{},
		resyncPeriod,
		indexers,
	)
}

func (f *certificateInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCertificateInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *certificateInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&certmanager_k8s_io_v1alpha2.Certificate{}, f.defaultInformer)
}

func (f *certificateInformer) Lister() v1alpha2.CertificateLister {
	return v1alpha2.NewCertificateLister(f.Informer().GetIndexer())
}
