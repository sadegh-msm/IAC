package controller

import (
	"context"
	"fmt"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MongoDBClusterReconciler) buildShardStatefulSet(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster, shardIndex int) error {
	labels := map[string]string{
		"app":   fmt.Sprintf("mongo-shard%d", shardIndex),
	}
	name := fmt.Sprintf("mongo-shard%d", shardIndex)
	serviceName := fmt.Sprintf("%s-%s%d", mongo.Name, "shard", shardIndex)

	obj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: mongo.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    pointer.Int32(int32(mongo.Spec.ReplicaSetSize)),
			ServiceName: serviceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "mongod",
						Image: "mongo:" + mongo.Spec.Version,
						Args: []string{
							"mongod",
							"--replSet", name,
							"--shardsvr",
							"--bind_ip_all",
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 27018,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "data",
							MountPath: "/data/db",
						}},
					}},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(mongo.Spec.StorageSize),
						},
					},
				},
			}},
		},
	}

	if err := ctrl.SetControllerReference(mongo, obj, r.Scheme); err != nil {
		return err
	}

	var existing appsv1.StatefulSet
	err := r.Get(ctx, client.ObjectKey{Name: obj.Name, Namespace: obj.Namespace}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, obj)
		}
		return err
	}

	obj.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, obj)
}

func (r *MongoDBClusterReconciler) buildConfigServerStatefulSet(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster) error {
	obj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "config",
			Namespace: mongo.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: mongo.Name + "-configsvr",
			Replicas:    pointer.Int32(int32(mongo.Spec.ConfigServerCount)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": mongo.Name + "-configsvr",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": mongo.Name + "-configsvr",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "mongod",
						Image: "mongo:" + mongo.Spec.Version,
						Args: []string{
							"mongod",
							"--configsvr",
							"--replSet", "configReplSet",
							"--bind_ip_all",
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 27019,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "data",
							MountPath: "/data/db",
						}},
					}},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(mongo.Spec.StorageSize),
						},
					},
				},
			}},
		},
	}

	if err := ctrl.SetControllerReference(mongo, obj, r.Scheme); err != nil {
		return err
	}

	var existing appsv1.StatefulSet
	err := r.Get(ctx, client.ObjectKey{Name: obj.Name, Namespace: obj.Namespace}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, obj)
		}
		return err
	}

	obj.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, obj)
}
