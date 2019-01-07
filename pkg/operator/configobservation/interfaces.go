package configobservation

import (
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
)

type Listers struct {
	ImageConfigLister configlistersv1.ImageLister
	AuthConfigLister  configlistersv1.AuthenticationLister
	EndpointsLister   corelistersv1.EndpointsLister
	ConfigmapLister   corelistersv1.ConfigMapLister

	ImageConfigSynced cache.InformerSynced
	AuthConfigSynced  cache.InformerSynced

	PreRunCachesSynced []cache.InformerSynced
}

func (l Listers) PreRunHasSynced() []cache.InformerSynced {
	return l.PreRunCachesSynced
}
