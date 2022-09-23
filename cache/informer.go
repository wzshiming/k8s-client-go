package cache

import (
	"context"
	"reflect"
	"time"

	"github.com/wzshiming/k8s-client-go/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type ResourceEventHandler[T client.Object] interface {
	OnAdd(obj T)
	OnUpdate(oldObj, newObj T)
	OnDelete(obj T)
}

type Store[T client.Object] interface {
	// Add adds the given object to the accumulator associated with the given object's key
	Add(obj T) error

	// Update updates the given object in the accumulator associated with the given object's key
	Update(obj T) error

	// Delete deletes the given object from the accumulator associated with the given object's key
	Delete(obj T) error

	// List returns a list of all the currently non-empty accumulators
	List() []T

	// ListKeys returns a list of all the keys currently associated with non-empty accumulators
	ListKeys() []string

	// Get returns the accumulator associated with the given object's key
	Get(obj T) (item T, exists bool, err error)

	// GetByKey returns the accumulator associated with the given key
	GetByKey(key string) (item T, exists bool, err error)

	// Replace will delete the contents of the store, using instead the
	// given list. Store takes ownership of the list, you should not reference
	// it after calling this function.
	Replace([]T, string) error

	// Resync is meaningless in the terms appearing here but has
	// meaning in some implementations that have non-trivial
	// additional behavior (e.g., DeltaFIFO).
	Resync() error
}

type store[T client.Object] struct {
	cache.Store
}

func (s store[T]) Add(obj T) error {
	return s.Store.Add(obj)
}

// Update updates the given object in the accumulator associated with the given object's key
func (s store[T]) Update(obj T) error {
	return s.Store.Update(obj)
}

// Delete deletes the given object from the accumulator associated with the given object's key
func (s store[T]) Delete(obj T) error {
	return s.Store.Delete(obj)
}

// List returns a list of all the currently non-empty accumulators
func (s store[T]) List() []T {
	items := s.Store.List()
	list := make([]T, 0, len(items))
	for _, item := range items {
		list = append(list, item.(T))
	}
	return list
}

// Get returns the accumulator associated with the given object's key
func (s store[T]) Get(obj T) (item T, exists bool, err error) {
	i, exists, err := s.Store.Get(obj)
	if i != nil {
		item = i.(T)
	}
	return item, exists, err
}

// GetByKey returns the accumulator associated with the given key
func (s store[T]) GetByKey(key string) (item T, exists bool, err error) {
	i, exists, err := s.Store.GetByKey(key)
	if i != nil {
		item = i.(T)
	}
	return item, exists, err
}

// Replace will delete the contents of the store, using instead the
// given list. Store takes ownership of the list, you should not reference
// it after calling this function.
func (s store[T]) Replace(items []T, resourceVersion string) error {
	in := make([]any, 0, len(items))
	for _, item := range items {
		in = append(in, item)
	}
	return s.Store.Replace(in, resourceVersion)
}

type Controller = cache.Controller

func Informer[T client.Object, L client.List](ctx context.Context, c client.Interface[T, L], resyncPeriod time.Duration, h ResourceEventHandler[T], optionsModifier func(options *metav1.ListOptions)) (Store[T], Controller) {
	var handler cache.ResourceEventHandler
	if h != nil {
		handler = cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				h.OnAdd(obj.(T))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				h.OnUpdate(oldObj.(T), newObj.(T))
			},
			DeleteFunc: func(obj interface{}) {
				h.OnDelete(obj.(T))
			},
		}
	}
	var (
		t T
	)
	t = reflect.New(reflect.TypeOf(t).Elem()).Interface().(T)

	s, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if optionsModifier != nil {
					optionsModifier(&options)
				}
				return c.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if optionsModifier != nil {
					optionsModifier(&options)
				}
				return c.Watch(ctx, options)
			},
		},
		t,
		resyncPeriod,
		handler,
	)
	return store[T]{s}, controller
}
