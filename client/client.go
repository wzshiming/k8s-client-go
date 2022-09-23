package client

import (
	"context"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type Interface[T Object, L List] interface {
	// Get takes name of the resource, and returns the corresponding resource object, and an error if there is any.
	Get(ctx context.Context, name string, options metav1.GetOptions) (result T, err error)

	// List takes label and field selectors, and returns the list of resource that match those selectors.
	List(ctx context.Context, opts metav1.ListOptions) (result L, err error)

	// Watch returns a watch.Interface that watches the requested restClient.
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)

	// Create takes the representation of a resource and creates it.  Returns the server's representation of the resource, and an error, if there is any.
	Create(ctx context.Context, cr T, opts metav1.CreateOptions) (result T, err error)

	// Update takes the representation of a resource and updates it. Returns the server's representation of the resource, and an error, if there is any.
	Update(ctx context.Context, cr T, opts metav1.UpdateOptions) (result T, err error)

	// UpdateStatus was generated because the type contains a Status member.
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, cr T, opts metav1.UpdateOptions) (result T, err error)

	// Delete takes name of the resource and deletes it. Returns an error if one occurs.
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error

	// DeleteCollection deletes a collection of objects.
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error

	// Patch applies the patch and returns the patched resource.
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result T, err error)
}

type Object interface {
	runtime.Object
	metav1.Object
}

type List interface {
	runtime.Object
	metav1.ListMetaAccessor
}

// client implements client[Object, List]
type client[T Object, L List] struct {
	scheme         *runtime.Scheme
	resource       string
	parameterCodec runtime.ParameterCodec
	restClient     rest.Interface
	ns             string

	tType reflect.Type
	lType reflect.Type
}

// NewClient returns a Interface[T Object, L List]
func NewClient[T Object, L List](scheme *runtime.Scheme, parameterCodec runtime.ParameterCodec, restClient rest.Interface, resource, namespace string) Interface[T, L] {
	var (
		t T
		l L
	)

	return &client[T, L]{
		scheme:         scheme,
		parameterCodec: parameterCodec,
		restClient:     restClient,
		resource:       resource,
		ns:             namespace,
		tType:          reflect.TypeOf(t).Elem(),
		lType:          reflect.TypeOf(l).Elem(),
	}
}

func (c *client[T, L]) newT() (result T) {
	return reflect.New(c.tType).Interface().(T)
}

func (c *client[T, L]) newL() (result L) {
	return reflect.New(c.lType).Interface().(L)
}

// Get takes name of the resource, and returns the corresponding resource object, and an error if there is any.
func (c *client[T, L]) Get(ctx context.Context, name string, options metav1.GetOptions) (result T, err error) {
	result = c.newT()
	err = c.restClient.Get().
		Namespace(c.ns).
		Resource(c.resource).
		Name(name).
		VersionedParams(&options, c.parameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of resource that match those selectors.
func (c *client[T, L]) List(ctx context.Context, opts metav1.ListOptions) (result L, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = c.newL()
	err = c.restClient.Get().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, c.parameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested restClient.
func (c *client[T, L]) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.restClient.Get().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, c.parameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a resource and creates it.  Returns the server's representation of the resource, and an error, if there is any.
func (c *client[T, L]) Create(ctx context.Context, cr T, opts metav1.CreateOptions) (result T, err error) {
	result = c.newT()
	err = c.restClient.Post().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&opts, c.parameterCodec).
		Body(cr).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a resource and updates it. Returns the server's representation of the resource, and an error, if there is any.
func (c *client[T, L]) Update(ctx context.Context, cr T, opts metav1.UpdateOptions) (result T, err error) {
	result = c.newT()
	err = c.restClient.Put().
		Namespace(c.ns).
		Resource(c.resource).
		Name(cr.GetName()).
		VersionedParams(&opts, c.parameterCodec).
		Body(cr).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *client[T, L]) UpdateStatus(ctx context.Context, cr T, opts metav1.UpdateOptions) (result T, err error) {
	result = c.newT()
	err = c.restClient.Put().
		Namespace(c.ns).
		Resource(c.resource).
		Name(cr.GetName()).
		SubResource("status").
		VersionedParams(&opts, c.parameterCodec).
		Body(cr).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the resource and deletes it. Returns an error if one occurs.
func (c *client[T, L]) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.Delete().
		Namespace(c.ns).
		Resource(c.resource).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *client[T, L]) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.restClient.Delete().
		Namespace(c.ns).
		Resource(c.resource).
		VersionedParams(&listOpts, c.parameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched resource.
func (c *client[T, L]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result T, err error) {
	result = c.newT()
	err = c.restClient.Patch(pt).
		Namespace(c.ns).
		Resource(c.resource).
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, c.parameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
