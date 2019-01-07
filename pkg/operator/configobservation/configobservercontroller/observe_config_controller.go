package configobservercontroller

import (
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"

	kubeapiserveroperatorinformers "github.com/openshift/cluster-kube-apiserver-operator/pkg/generated/informers/externalversions"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation/auth"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation/etcd"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation/images"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation/network"
)

type ConfigObserver struct {
	*configobserver.ConfigObserver
}

func NewConfigObserver(
	operatorClient v1helpers.OperatorClient,
	operatorConfigInformers kubeapiserveroperatorinformers.SharedInformerFactory,
	kubeInformersForKubeSystemNamespace kubeinformers.SharedInformerFactory,
	configInformer configinformers.SharedInformerFactory,
	eventRecorder events.Recorder,
) *ConfigObserver {
	c := &ConfigObserver{
		ConfigObserver: configobserver.NewConfigObserver(
			operatorClient,
			eventRecorder,
			configobservation.Listers{
				AuthConfigLister:  configInformer.Config().V1().Authentications().Lister(),
				ImageConfigLister: configInformer.Config().V1().Images().Lister(),
				EndpointsLister:   kubeInformersForKubeSystemNamespace.Core().V1().Endpoints().Lister(),
				ConfigmapLister:   kubeInformersForKubeSystemNamespace.Core().V1().ConfigMaps().Lister(),
				ImageConfigSynced: configInformer.Config().V1().Images().Informer().HasSynced,
				AuthConfigSynced:  configInformer.Config().V1().Authentications().Informer().HasSynced,
				PreRunCachesSynced: []cache.InformerSynced{
					operatorConfigInformers.Kubeapiserver().V1alpha1().KubeAPIServerOperatorConfigs().Informer().HasSynced,
					kubeInformersForKubeSystemNamespace.Core().V1().Endpoints().Informer().HasSynced,
					kubeInformersForKubeSystemNamespace.Core().V1().ConfigMaps().Informer().HasSynced,
				},
			},
			etcd.ObserveStorageURLs,
			network.ObserveRestrictedCIDRs,
			images.ObserveInternalRegistryHostname,
			images.ObserveExternalRegistryHostnames,
			images.ObserveAllowedRegistriesForImport,
			auth.ObserveAuthMetadata,
		),
	}

	operatorConfigInformers.Kubeapiserver().V1alpha1().KubeAPIServerOperatorConfigs().Informer().AddEventHandler(c.EventHandler())
	kubeInformersForKubeSystemNamespace.Core().V1().Endpoints().Informer().AddEventHandler(c.EventHandler())
	kubeInformersForKubeSystemNamespace.Core().V1().ConfigMaps().Informer().AddEventHandler(c.EventHandler())
	configInformer.Config().V1().Images().Informer().AddEventHandler(c.EventHandler())
	configInformer.Config().V1().Authentications().Informer().AddEventHandler(c.EventHandler())

	return c
}
