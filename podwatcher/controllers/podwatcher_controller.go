/*
Copyright 2023.

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

package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	sregigv1alpha1 "github.com/tokalevasant/sre-gig-tasks/podwatcher/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodwatcherReconciler reconciles a Podwatcher object
type PodwatcherReconciler struct {
	Log logr.Logger
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=sregig.redhat.com,resources=podwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sregig.redhat.com,resources=podwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sregig.redhat.com,resources=podwatchers/finalizers,verbs=update;
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Podwatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *PodwatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the PodWatcher Instance
	podwatcher := &sregigv1alpha1.Podwatcher{}

	err := r.Get(ctx, req.NamespacedName, podwatcher)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("1.  Fetch the Podwatcher instance. Podwatcher resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "1.  Fetch the Podwatcher instance. Failed to get Podwatcher")
		return ctrl.Result{}, nil
	}

	logger.Info("1. Fetch the Podwatcher instance. ", "podwatcher.Name", podwatcher.Name, "podwatcher.Namespace", podwatcher.Namespace)
	// 2. Check if the deployment exists, if not create one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: podwatcher.Name, Namespace: podwatcher.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForPodWatcher(podwatcher)
		logger.Info("dep", ":", *dep)
		logger.Info("2. Check if the deployment already exists, if not create a new one. Creating new Deployment", "Deployment.Namespace", podwatcher.Namespace, "Deployment.Name", podwatcher.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error(err, "2. Check if the deplouyment already exists, if not create a new one. Failed to create new Deployment.", "Deployment.Namespace", podwatcher.Namespace, "Deployment.Name", podwatcher.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "2. Check if the deplouyment already exists, if not create a new one. Failed to get Deployment.", "Deployment.Namespace", podwatcher.Namespace, "Deployment.Name", podwatcher.Name)
		return ctrl.Result{}, err
	}

	// 3. Match the Deployment & the spec
	size := podwatcher.Spec.Size
	logger.Info("+++++++++", "replicaa", *found)
	if *found.Spec.Replicas != size {

		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			logger.Error(err, "3. Ensure the deployment size is the same as the spec. Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}

		logger.Info("3. Ensure the deployment size is the same as the spec. Update deployment size", "Deployment.Spec.Replicas", size)
		return ctrl.Result{}, nil
	}

	// 4. Print the pod names to the logs i.e. main logic
	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(podwatcher.Namespace),
		client.MatchingLabels(r.labelsForPodWatcher(podwatcher.Name)),
	}

	if err = r.List(ctx, podlist, listOpts...); err != nil {
		logger.Error(err, "4. Update the PodWatcher status with the pod names. Failed to list pods.", "PodWatcher.Namespace", podwatcher.Namespace, "PodWatcher.Name", podwatcher.Name)
		return ctrl.Result{}, err
	}

	podNames := getPodNames(podlist.Items)

	logger.Info("4. Update the PodWatcher status with the pod names. Pod list", "podnames", podNames)

	if !reflect.DeepEqual(podNames, podwatcher.Status.Pods) {
		podwatcher.Status.Pods = podNames
		err := r.Status().Update(ctx, podwatcher)
		if err != nil {
			logger.Error(err, "4. Update the PodWatcher status with the pod names. Failed to update PodWatcher status.", "PodWatcher.Status.Pods", podwatcher.Status.Pods, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodwatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sregigv1alpha1.Podwatcher{}).
		Complete(r)
}

func (r *PodwatcherReconciler) deploymentForPodWatcher(m *sregigv1alpha1.Podwatcher) *appsv1.Deployment {

	ls := r.labelsForPodWatcher(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "podwatcher:latest",
						Name:    "podwatcher",
						Command: []string{"podwatcher", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "podwatcher",
						}},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *PodwatcherReconciler) labelsForPodWatcher(name string) map[string]string {
	return map[string]string{"app": "podwatcher", "podwatcher": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
