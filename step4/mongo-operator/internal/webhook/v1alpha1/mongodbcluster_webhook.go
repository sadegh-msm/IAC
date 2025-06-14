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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var mongodbclusterlog = logf.Log.WithName("mongodbcluster-resource")

// SetupMongoDBClusterWebhookWithManager registers the webhook for MongoDBCluster in the manager.
func SetupMongoDBClusterWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&databasev1alpha1.MongoDBCluster{}).
		WithValidator(&MongoDBClusterCustomValidator{}).
		WithDefaulter(&MongoDBClusterCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-database-sadegh-msm-v1alpha1-mongodbcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=database.sadegh.msm,resources=mongodbclusters,verbs=create;update,versions=v1alpha1,name=mmongodbcluster-v1alpha1.kb.io,admissionReviewVersions=v1

// MongoDBClusterCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind MongoDBCluster when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type MongoDBClusterCustomDefaulter struct{}

var _ webhook.CustomDefaulter = &MongoDBClusterCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind MongoDBCluster.
func (d *MongoDBClusterCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	mongodbcluster, ok := obj.(*databasev1alpha1.MongoDBCluster)

	if !ok {
		return fmt.Errorf("expected an MongoDBCluster object but got %T", obj)
	}
	mongodbclusterlog.Info("Defaulting for MongoDBCluster", "name", mongodbcluster.GetName())

	// Set default number of replicas if not set
	if mongodbcluster.Spec.ReplicaSetSize == 0 {
		mongodbclusterlog.Info("Defaulting replicas to 3")
		mongodbcluster.Spec.ReplicaSetSize = 3
	}

	// Set default MongoDB version if not set
	if mongodbcluster.Spec.Version == "" {
		mongodbclusterlog.Info("Defaulting version to 7")
		mongodbcluster.Spec.Version = "7"
	}

	// Set default backup schedule if enabled but no schedule provided
	if mongodbcluster.Spec.Backup.Enabled && mongodbcluster.Spec.Backup.Schedule == "" {
		mongodbclusterlog.Info("Defaulting backup schedule to '0 2 * * *'")
		mongodbcluster.Spec.Backup.Schedule = "0 2 * * *"
	}

	// Set default backup secret namespace if not specified
	if mongodbcluster.Spec.Backup.SecretRef.Name != "" && mongodbcluster.Spec.Backup.SecretRef.Namespace == "" {
		mongodbclusterlog.Info("Defaulting SecretRef namespace to resource namespace")
		mongodbcluster.Spec.Backup.SecretRef.Namespace = mongodbcluster.Namespace
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-database-sadegh-msm-v1alpha1-mongodbcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=database.sadegh.msm,resources=mongodbclusters,verbs=create;update,versions=v1alpha1,name=vmongodbcluster-v1alpha1.kb.io,admissionReviewVersions=v1

// MongoDBClusterCustomValidator struct is responsible for validating the MongoDBCluster resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type MongoDBClusterCustomValidator struct{}

var _ webhook.CustomValidator = &MongoDBClusterCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type MongoDBCluster.
func (v *MongoDBClusterCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	mongodbcluster, ok := obj.(*databasev1alpha1.MongoDBCluster)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDBCluster object but got %T", obj)
	}
	mongodbclusterlog.Info("Validation for MongoDBCluster upon creation", "name", mongodbcluster.GetName())

	return nil, validateSpec(mongodbcluster)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type MongoDBCluster.
func (v *MongoDBClusterCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	mongodbcluster, ok := newObj.(*databasev1alpha1.MongoDBCluster)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDBCluster object for the newObj but got %T", newObj)
	}
	mongodbclusterlog.Info("Validation for MongoDBCluster upon update", "name", mongodbcluster.GetName())

	return nil, validateSpec(mongodbcluster)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type MongoDBCluster.
func (v *MongoDBClusterCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	mongodbcluster, ok := obj.(*databasev1alpha1.MongoDBCluster)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDBCluster object but got %T", obj)
	}
	mongodbclusterlog.Info("Validation for MongoDBCluster upon deletion", "name", mongodbcluster.GetName())

	return nil, nil
}

// validateSpec checks backup and sharding constraints
func validateSpec(m *databasev1alpha1.MongoDBCluster) error {
	var errs []string

	// Backup validation
	if m.Spec.Backup.Enabled {
		if m.Spec.Backup.Schedule == "" {
			errs = append(errs, "spec.backup.schedule cannot be empty")
		}
		if m.Spec.Backup.StorageEndpoint == "" {
			errs = append(errs, "spec.backup.storageEndpoint cannot be empty")
		}
		if m.Spec.Backup.Bucket == "" {
			errs = append(errs, "spec.backup.bucket cannot be empty")
		}
		if m.Spec.Backup.SecretRef.Name == "" || m.Spec.Backup.SecretRef.Namespace == "" {
			errs = append(errs, "spec.backup.secretRef.name or namespace cannot be empty")
		}
	}

	if m.Spec.StorageSize == "" {
		errs = append(errs, "spec.storageSize is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("spec validation failed: %v", errs)
	}

	return nil
}
