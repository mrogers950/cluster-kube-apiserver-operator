package auth

import (
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/openshift/library-go/pkg/operator/configobserver"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-kube-apiserver-operator/pkg/operator/configobservation"
)

//const clusterAuthConfigNamespace = "openshift-config-managed"

// ObserveInternalRegistryHostname reads the internal registry hostname from the cluster configuration as provided by
// the registry operator.
func ObserveAuthMetadata(genericListers configobserver.Listers, recorder events.Recorder, existingConfig map[string]interface{}) (map[string]interface{}, []error) {
	listers := genericListers.(configobservation.Listers)
	errs := []error{}
	prevObservedConfig := map[string]interface{}{}

	topLevelMetadataFilePath := []string{"authConfig", "oauthMetadataFile"}
	currentMetadataFilePath, _, err := unstructured.NestedString(existingConfig, topLevelMetadataFilePath...)
	if err != nil {
		errs = append(errs, err)
	}
	if len(currentMetadataFilePath) > 0 {
		if err := unstructured.SetNestedField(prevObservedConfig, currentMetadataFilePath, topLevelMetadataFilePath...); err != nil {
			errs = append(errs, err)
		}
	}

	if !listers.AuthConfigSynced() {
		glog.Warning("authentications.config.openshift.io not synced")
		return prevObservedConfig, errs
	}

	observedConfig := map[string]interface{}{}
	authConfig, err := listers.AuthConfigLister.Get("cluster")
	if errors.IsNotFound(err) {
		glog.Warningf("authentications.config.openshift.io/cluster: not found")
		return observedConfig, errs
	}
	if err != nil {
		glog.Warningf("DBG: err getting authentications.config.openshift.io/cluster %v", err)
		return prevObservedConfig, errs
	}

	var oauthMetadataConfigMap v1.ConfigMapReference
	oauthMetadataConfigMap = authConfig.Spec.OAuthMetadata
	if len(oauthMetadataConfigMap.Name) == 0 || len(oauthMetadataConfigMap.Key) == 0 {
		glog.Warningf("DBG: using status metadata: %s/%s", authConfig.Status.OAuthMetadata.Name, authConfig.Status.OAuthMetadata.Key)
		oauthMetadataConfigMap = authConfig.Status.OAuthMetadata
	}

	metadataConfigMap, err := listers.ConfigmapLister.ConfigMaps("kube-system").Get(oauthMetadataConfigMap.Name)
	if err != nil {
		recorder.Eventf("ObserveAuthMetadataConfigMap", "Failed to get oauth metadata configMap %s/%s: %v", "kube-system", oauthMetadataConfigMap.Name, err)
		return prevObservedConfig, errs
	}

	if _, ok := metadataConfigMap.Data[oauthMetadataConfigMap.Key]; ok {
		glog.Warningf("DBG: got confmap key %s, updating", oauthMetadataConfigMap.Key)

		if len(currentMetadataFilePath) == 0 {
			glog.Warningf("DBG: metadata path empty, setting")
			if err := unstructured.SetNestedField(observedConfig, "/etc/kubernetes/static-pod-resources/configmaps/oauth-metadata/"+oauthMetadataConfigMap.Key, topLevelMetadataFilePath...); err != nil {
				glog.Warningf("DBG: error setting metadata: %v", err)
				errs = append(errs, err)
			}
		}
	} else {
		glog.Warningf("DBG: no confmap key %s, unsetting", oauthMetadataConfigMap.Key)

		if len(currentMetadataFilePath) > 0 {
			glog.Warningf("DBG: metadata was set, unsetting")

			if err := unstructured.SetNestedField(observedConfig, "", topLevelMetadataFilePath...); err != nil {
				glog.Warningf("DBG: error setting metadata: %v", err)
				errs = append(errs, err)
			}
		}
	}
	//if len(internalRegistryHostName) > 0 {
	//	if err := unstructured.SetNestedField(observedConfig, internalRegistryHostName, internalRegistryHostnamePath...); err != nil {
	//		errs = append(errs, err)
	//	}
	//	if internalRegistryHostName != currentInternalRegistryHostname {
	//		recorder.Eventf("ObserveRegistryHostnameChanged", "Internal registry hostname changed to %q", internalRegistryHostName)
	//	}
	//}
	return observedConfig, errs
}
