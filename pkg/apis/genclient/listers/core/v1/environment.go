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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/srcmesh/kubefaas/pkg/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// EnvironmentLister helps list Environments.
type EnvironmentLister interface {
	// List lists all Environments in the indexer.
	List(selector labels.Selector) (ret []*v1.Environment, err error)
	// Environments returns an object that can list and get Environments.
	Environments(namespace string) EnvironmentNamespaceLister
	EnvironmentListerExpansion
}

// _environmentLister implements the EnvironmentLister interface.
type _environmentLister struct {
	indexer cache.Indexer
}

// NewEnvironmentLister returns a new EnvironmentLister.
func NewEnvironmentLister(indexer cache.Indexer) EnvironmentLister {
	return &_environmentLister{indexer: indexer}
}

// List lists all Environments in the indexer.
func (s *_environmentLister) List(selector labels.Selector) (ret []*v1.Environment, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Environment))
	})
	return ret, err
}

// Environments returns an object that can list and get Environments.
func (s *_environmentLister) Environments(namespace string) EnvironmentNamespaceLister {
	return _environmentNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// EnvironmentNamespaceLister helps list and get Environments.
type EnvironmentNamespaceLister interface {
	// List lists all Environments in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.Environment, err error)
	// Get retrieves the Environment from the indexer for a given namespace and name.
	Get(name string) (*v1.Environment, error)
	EnvironmentNamespaceListerExpansion
}

// _environmentNamespaceLister implements the EnvironmentNamespaceLister
// interface.
type _environmentNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Environments in the indexer for a given namespace.
func (s _environmentNamespaceLister) List(selector labels.Selector) (ret []*v1.Environment, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Environment))
	})
	return ret, err
}

// Get retrieves the Environment from the indexer for a given namespace and name.
func (s _environmentNamespaceLister) Get(name string) (*v1.Environment, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("environment"), name)
	}
	return obj.(*v1.Environment), nil
}
