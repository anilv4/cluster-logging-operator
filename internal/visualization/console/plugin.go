package console

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcilePlugin reconciles the console plugin to expose log querying of storage
func ReconcilePlugin(k8sClient client.Client, cl *logging.ClusterLogging, owner client.Object, clusterVersion string) error {
	lokiService := lokiServiceName(cl)

	var consoleSpec *logging.OCPConsoleSpec
	if cl != nil && cl.Spec.Visualization != nil {
		consoleSpec = cl.Spec.Visualization.OCPConsole
	}

	r := NewReconciler(k8sClient, NewConfig(owner, lokiService, FeaturesForOCP(clusterVersion)), consoleSpec)
	if lokiService != "" {
		log.V(3).Info("Enabling logging console plugin", "created-by", r.CreatedBy(), "loki-service", lokiService)
		return r.Reconcile(context.TODO())
	} else {
		log.V(3).Info("Removing logging console plugin", "created-by", r.CreatedBy(), "loki-service", lokiService)
		return r.Delete(context.TODO())
	}
}

func lokiServiceName(cl *logging.ClusterLogging) string {
	if cl.Spec.LogStore != nil && cl.Spec.LogStore.Type == logging.LogStoreTypeLokiStack {
		return lokistack.LokiStackGatewayService(cl.Spec.LogStore)
	}

	if stackName, ok := cl.Annotations[constants.AnnotationOCPConsoleMigrationTarget]; ok {
		return lokistack.LokiStackGatewayService(&logging.LogStoreSpec{
			LokiStack: logging.LokiStackStoreSpec{
				Name: stackName,
			},
		})
	}

	return ""
}
