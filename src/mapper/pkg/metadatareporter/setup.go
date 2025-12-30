package metadatareporter

import (
	"context"
	"github.com/otterize/intents-operator/src/shared/errors"
	"github.com/otterize/intents-operator/src/shared/serviceidresolver"
	"github.com/otterize/network-mapper/src/mapper/pkg/cloudclient"
	discoveryv1 "k8s.io/api/discovery/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Setup(client client.Client, cloudClient cloudclient.CloudClient, resolver serviceidresolver.ServiceResolver, mgr ctrl.Manager) error {
	reporter := NewMetadataReporter(client, cloudClient, resolver)

	// Initialize indexes
	if err := initIndexes(mgr); err != nil {
		return errors.Wrap(err)
	}

	// Initialize the EndpointsReconciler
	endpointsReconciler := NewEndpointsReconciler(client, resolver, reporter)
	if err := endpointsReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err)
	}

	// Initialize the PodReconciler
	podReconciler := NewPodReconciler(client, resolver, reporter)
	if err := podReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err)
	}

	// Initialize the NamespaceReconciler
	namespaceReconciler := NewNamespaceReconciler(mgr.GetClient(), cloudClient)
	if err := namespaceReconciler.SetupWithManager(mgr); err != nil {
		return errors.Wrap(err)
	}

	return nil
}
func initIndexes(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&discoveryv1.EndpointSlice{},
		endpointsPodNamesIndexField,
		func(object client.Object) []string {
			var res []string
			endpointSlice := object.(*discoveryv1.EndpointSlice)

			for _, endpoint := range endpointSlice.Endpoints {
				if endpoint.TargetRef == nil || endpoint.TargetRef.Kind != "Pod" {
					continue
				}

				res = append(res, endpoint.TargetRef.Name)
			}

			return res
		}); err != nil {
		return errors.Wrap(err)
	}
	return nil
}
