/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
)

// MongoDBClusterReconciler reconciles a MongoDBCluster object
type MongoDBClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.sadegh.msm,resources=mongodbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.sadegh.msm,resources=mongodbclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.sadegh.msm,resources=mongodbclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MongoDBCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *MongoDBClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var mongo databasev1alpha1.MongoDBCluster

	if err := r.Get(ctx, req.NamespacedName, &mongo); err != nil {
		log.Error(err, "unable to fetch MongoDBCluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or update core components
	if err := r.buildConfigServerStatefulSet(ctx, &mongo); err != nil {
		log.Error(err, "failed to reconcile config server")
		return ctrl.Result{}, err
	}

	for i := 0; i < mongo.Spec.ReplicaSetCount; i++ {
		if err := r.reconcileShardService(ctx, &mongo, i); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.buildShardStatefulSet(ctx, &mongo, i); err != nil {
			log.Error(err, "failed to reconcile shard", "shard", i)
			return ctrl.Result{}, err
		}

		// if err := r.waitForAllReplicas(ctx, &mongo, i); err != nil {
		// 	return ctrl.Result{}, err
		// }
	}

	if err := r.buildMongosDeployment(ctx, &mongo); err != nil {
		log.Error(err, "failed to reconcile mongos")
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigSvrService(ctx, &mongo); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMongosService(ctx, &mongo); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigServerReplicaSet(ctx, &mongo); err != nil {
		return ctrl.Result{}, err
	}

	for i := 0; i < mongo.Spec.ReplicaSetCount; i++ {
		if err := r.reconcileReplicaSet(ctx, &mongo, i); err != nil {
			return ctrl.Result{}, err
		}
	}

	for i := 0; i < mongo.Spec.ReplicaSetCount; i++ {
		if err := r.reconcileEnbaleShardingMongos(ctx, &mongo, i); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileBackupCronJob(ctx, &mongo); err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	mongo.Status.Phase = "Running"
	totalShards := mongo.Spec.ReplicaSetCount * mongo.Spec.ReplicaSetSize
	mongo.Status.ReplicaSummary = fmt.Sprintf("%d/%d", totalShards, totalShards)
	mongo.Status.CurrentVersion = mongo.Spec.Version

	if err := r.Status().Update(ctx, &mongo); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoDBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1alpha1.MongoDBCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&batchv1.CronJob{}).
		Named("mongodbcluster").
		Complete(r)
}
