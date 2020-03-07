/*
Copyright The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	corev1 "github.com/srcmesh/kubefaas/pkg/apis/core/v1"
	versioned "github.com/srcmesh/kubefaas/pkg/apis/genclient/clientset/versioned"
	internalinterfaces "github.com/srcmesh/kubefaas/pkg/apis/genclient/informers/externalversions/internalinterfaces"
	v1 "github.com/srcmesh/kubefaas/pkg/apis/genclient/listers/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MessageQueueTriggerInformer provides access to a shared informer and lister for
// MessageQueueTriggers.
type MessageQueueTriggerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.MessageQueueTriggerLister
}

type _messageQueueTriggerInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMessageQueueTriggerInformer constructs a new informer for MessageQueueTrigger type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMessageQueueTriggerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMessageQueueTriggerInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMessageQueueTriggerInformer constructs a new informer for MessageQueueTrigger type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMessageQueueTriggerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().MessageQueueTriggers(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().MessageQueueTriggers(namespace).Watch(options)
			},
		},
		&corev1.MessageQueueTrigger{},
		resyncPeriod,
		indexers,
	)
}

func (f *_messageQueueTriggerInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMessageQueueTriggerInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *_messageQueueTriggerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&corev1.MessageQueueTrigger{}, f.defaultInformer)
}

func (f *_messageQueueTriggerInformer) Lister() v1.MessageQueueTriggerLister {
	return v1.NewMessageQueueTriggerLister(f.Informer().GetIndexer())
}
