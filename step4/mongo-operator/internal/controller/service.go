package controller

import (
	"context"
	"fmt"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mongoDBPort      = 27017
	mongoShardPort   = 27018
	configServerPort = 27019
)

func (r *MongoDBClusterReconciler) reconcileShardService(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster, shardIndex int) error {
	name := fmt.Sprintf("%s-shard%d", mongo.Name, shardIndex)
	labels := map[string]string{
		"app": fmt.Sprintf("mongo-shard%d", shardIndex),
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: mongo.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone, // Headless service
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "mongodb",
					Port:     mongoShardPort,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(mongo, svc, r.Scheme); err != nil {
		return err
	}

	var existing corev1.Service
	err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, svc)
		}
		return err
	}

	return nil
}

func (r *MongoDBClusterReconciler) reconcileConfigSvrService(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongo.Name + "-configsvr",
			Namespace: mongo.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "configsvr",
				"app.kubernetes.io/instance":  mongo.Name,
				"app.kubernetes.io/component": "configsvr",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None", // Headless for replica set
			Selector: map[string]string{
				"app": mongo.Name + "-configsvr",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "mongodb",
					Port:     configServerPort,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(mongo, svc, r.Scheme); err != nil {
		return err
	}

	var existing corev1.Service
	err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, svc)
		}
		return err
	}
	svc.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, svc)
}

func (r *MongoDBClusterReconciler) reconcileMongosService(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongo.Name + "-mongos",
			Namespace: mongo.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "mongos",
				"app.kubernetes.io/instance":  mongo.Name,
				"app.kubernetes.io/component": "mongos",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": mongo.Name + "-mongos",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "mongodb",
					Port:     mongoDBPort,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP, // Use NodePort or LoadBalancer if external access is needed
		},
	}

	if err := ctrl.SetControllerReference(mongo, svc, r.Scheme); err != nil {
		return err
	}

	var existing corev1.Service
	err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, svc)
		}
		return err
	}
	svc.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, svc)
}
