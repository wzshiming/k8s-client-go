package clientset

import (
	"github.com/wzshiming/k8s-client-go/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
}

func NewRestConfigFromKubeconfig(kubeconfig []byte) (*rest.Config, error) {
	cfg, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (conf *clientcmdapi.Config, err error) {
		return clientcmd.Load(kubeconfig)
	})
	if err != nil {
		return nil, err
	}
	err = setConfigDefaults(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewForConfig(restConfig *rest.Config) (Interface, error) {
	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, err
	}
	codecs := serializer.NewCodecFactory(scheme)
	restConfig.NegotiatedSerializer = codecs.WithoutConversion()

	restClient, err := rest.RESTClientForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, err
	}

	parameterCodec := runtime.NewParameterCodec(scheme)

	return &clientset{
		scheme:         scheme,
		parameterCodec: parameterCodec,
		client:         restClient,
	}, nil
}

func setConfigDefaults(config *rest.Config) error {
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	return rest.SetKubernetesDefaults(config)
}

type Interface interface {
	Namespaces() client.Interface[*corev1.Namespace, *corev1.NamespaceList]
	ConfigMaps(namespace string) client.Interface[*corev1.ConfigMap, *corev1.ConfigMapList]
	Secrets(namespace string) client.Interface[*corev1.Secret, *corev1.SecretList]
	Pods(namespace string) client.Interface[*corev1.Pod, *corev1.PodList]
	Services(namespace string) client.Interface[*corev1.Service, *corev1.ServiceList]
	Endpoints(namespace string) client.Interface[*corev1.Endpoints, *corev1.EndpointsList]
}

type clientset struct {
	scheme         *runtime.Scheme
	parameterCodec runtime.ParameterCodec
	client         rest.Interface
}

func (c *clientset) Namespaces() client.Interface[*corev1.Namespace, *corev1.NamespaceList] {
	return client.NewClient[*corev1.Namespace, *corev1.NamespaceList](c.scheme, c.parameterCodec, c.client, "namespaces", "")
}

func (c *clientset) ConfigMaps(namespace string) client.Interface[*corev1.ConfigMap, *corev1.ConfigMapList] {
	return client.NewClient[*corev1.ConfigMap, *corev1.ConfigMapList](c.scheme, c.parameterCodec, c.client, "configmaps", namespace)
}

func (c *clientset) Secrets(namespace string) client.Interface[*corev1.Secret, *corev1.SecretList] {
	return client.NewClient[*corev1.Secret, *corev1.SecretList](c.scheme, c.parameterCodec, c.client, "secrets", namespace)
}

func (c *clientset) Pods(namespace string) client.Interface[*corev1.Pod, *corev1.PodList] {
	return client.NewClient[*corev1.Pod, *corev1.PodList](c.scheme, c.parameterCodec, c.client, "pods", namespace)
}

func (c *clientset) Services(namespace string) client.Interface[*corev1.Service, *corev1.ServiceList] {
	return client.NewClient[*corev1.Service, *corev1.ServiceList](c.scheme, c.parameterCodec, c.client, "services", namespace)
}

func (c *clientset) Endpoints(namespace string) client.Interface[*corev1.Endpoints, *corev1.EndpointsList] {
	return client.NewClient[*corev1.Endpoints, *corev1.EndpointsList](c.scheme, c.parameterCodec, c.client, "endpoints", namespace)
}
