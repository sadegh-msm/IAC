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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MongoDBClusterSpec defines the desired state of MongoDBCluster.
type MongoDBClusterSpec struct {
	ReplicaSetCount   int           `json:"replicaSetCount"`
	ReplicaSetSize    int           `json:"replicaSetSize"`
	Sharding          *ShardingSpec `json:"sharding,omitempty"`
	ConfigServerCount int           `json:"configServerCount"`
	MongosCount       int           `json:"mongosCount"`
	Version           string        `json:"version"`
	Backup            BackupSpec    `json:"backup"`
	StorageSize       string        `json:"storageSize"`
	StorageClass      string        `json:"storageClass,omitempty"`
}

// ShardingSpec defines sharding options for a MongoDBCluster
type ShardingSpec struct {
	Enabled     bool   `json:"enabled"`
	Database    string `json:"database"`              // The database to enable sharding on
	Collections string `json:"collections,omitempty"` // The collections to shard
	Key         string `json:"key,omitempty"`
}

type BackupSpec struct {
	Enabled         bool      `json:"enabled"`
	Schedule        string    `json:"schedule,omitempty"` // Cron format
	StorageEndpoint string    `json:"storageEndpoint,omitempty"`
	Bucket          string    `json:"bucket,omitempty"`
	SecretRef       SecretRef `json:"secretRef,omitempty"`
}

type SecretRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// MongoDBClusterStatus defines the observed state of MongoDBCluster.
type MongoDBClusterStatus struct {
	Phase              string             `json:"phase,omitempty"`              // e.g., "Pending", "Running", "Degraded", "Failed"
	Message            string             `json:"message,omitempty"`            // Human-readable status
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`      // Number of ready MongoDB pods
	ReplicaSummary     string             `json:"replicaSummary,omitempty"`     // e.g., "3/3"
	CurrentVersion     string             `json:"currentVersion,omitempty"`     // Detected MongoDB version
	LastBackupTime     *metav1.Time       `json:"lastBackupTime,omitempty"`     // Last successful backup time
	ShardingConfigured bool               `json:"shardingConfigured,omitempty"` // Whether sharding has been successfully set up
	Conditions         []ClusterCondition `json:"conditions,omitempty"`         // List of status conditions
}

type ClusterCondition struct {
	Type               string                 `json:"type"`   // e.g., "ReplicaSetReady", "ShardingReady", "BackupSucceeded"
	Status             corev1.ConditionStatus `json:"status"` // "True", "False", "Unknown"
	LastTransitionTime metav1.Time            `json:"lastTransitionTime"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Cluster Phase"
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=".status.replicaSummary",description="Ready/Desired replicas"
// +kubebuilder:printcolumn:name="Sharding",type=boolean,JSONPath=".spec.sharding",description="Sharding Enabled"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=".spec.version",description="MongoDB Version"
// +kubebuilder:printcolumn:name="Last Backup",type=date,JSONPath=".status.lastBackupTime",description="Last successful backup time"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// MongoDBCluster is the Schema for the mongodbclusters API.
type MongoDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBClusterSpec   `json:"spec,omitempty"`
	Status MongoDBClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MongoDBClusterList contains a list of MongoDBCluster.
type MongoDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBCluster{}, &MongoDBClusterList{})
}
