package metadatareporter

import (
	"context"
	"github.com/otterize/intents-operator/src/shared/errors"
	"github.com/otterize/intents-operator/src/shared/serviceidresolver"
	"github.com/otterize/intents-operator/src/shared/serviceidresolver/serviceidentity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EndpointsReconciler struct {
	client.Client
	serviceIDResolver serviceidresolver.ServiceResolver
	metadataReporter  *MetadataReporter
}

func NewEndpointsReconciler(client client.Client, resolver serviceidresolver.ServiceResolver, reporter *MetadataReporter) *EndpointsReconciler {
	return &EndpointsReconciler{
		Client:            client,
		serviceIDResolver: resolver,
		metadataReporter:  reporter,
	}
}

func (r *EndpointsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&discoveryv1.EndpointSlice{}).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(r.mapServicesToEndpointSlices)).
		WithOptions(controller.Options{RecoverPanic: lo.ToPtr(true)}).
		Complete(r)
}

func (r *EndpointsReconciler) mapServicesToEndpointSlices(ctx context.Context, obj client.Object) []reconcile.Request {
	service := obj.(*corev1.Service)
	logrus.Debugf("Enqueueing endpoint slices for service %s", service.Name)

	// List all EndpointSlices for this service
	var endpointSlices discoveryv1.EndpointSliceList
	err := r.List(ctx, &endpointSlices, client.InNamespace(service.GetNamespace()), client.MatchingLabels{
		discoveryv1.LabelServiceName: service.GetName(),
	})
	if err != nil {
		logrus.WithError(err).Errorf("Failed to list endpoint slices for service %s", service.Name)
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, 0, len(endpointSlices.Items))
	for _, endpointSlice := range endpointSlices.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: endpointSlice.GetNamespace(),
				Name:      endpointSlice.GetName(),
			},
		})
	}

	return requests
}

func (r *EndpointsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	endpointSlice := &discoveryv1.EndpointSlice{}
	err := r.Get(ctx, req.NamespacedName, endpointSlice)
	if err != nil && client.IgnoreNotFound(err) == nil {
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err)
	}

	if endpointSlice.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	podNames := r.getPodNamesFromEndpointSlice(endpointSlice)
	serviceIdentities := make(map[string]serviceidentity.ServiceIdentity)
	for _, podName := range podNames {
		pod := &corev1.Pod{}
		err := r.Get(ctx, client.ObjectKey{Namespace: endpointSlice.Namespace, Name: podName}, pod)
		if err != nil && client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err)
		}
		serviceIdentity, err := r.serviceIDResolver.ResolvePodToServiceIdentity(ctx, pod)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err)
		}
		serviceIdentities[serviceIdentity.GetNameWithKind()] = serviceIdentity
	}

	if len(serviceIdentities) == 0 {
		return ctrl.Result{}, nil
	}

	err = r.metadataReporter.ReportMetadata(ctx, lo.Values(serviceIdentities))
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err)
	}

	return ctrl.Result{}, nil
}

func (r *EndpointsReconciler) getPodNamesFromEndpointSlice(endpointSlice *discoveryv1.EndpointSlice) []string {
	podNames := make([]string, 0)
	for _, endpoint := range endpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			podNames = append(podNames, endpoint.TargetRef.Name)
		}
	}
	return podNames
}
