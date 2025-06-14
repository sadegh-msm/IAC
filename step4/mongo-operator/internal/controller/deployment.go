package controller

import (
	"context"
	"fmt"
	"strings"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MongoDBClusterReconciler) buildMongosDeployment(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster) error {
	var hosts []string
	for i := 0; i < mongo.Spec.ConfigServerCount; i++ {
		host := fmt.Sprintf("config-%d.%s-configsvr.%s.svc.cluster.local:27019", i, mongo.Name, mongo.Namespace)
		hosts = append(hosts, host)
	}
	configDB := fmt.Sprintf("configReplSet/%s", joinHosts(hosts))

	labels := map[string]string{
		"app":       mongo.Name + "-mongos",
		"component": "mongos",
	}

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongos",
			Namespace: mongo.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(int32(mongo.Spec.MongosCount)),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "mongos",
						Image: "mongo:" + mongo.Spec.Version,
						Args: []string{
							"mongos",
							"--configdb=" + configDB,
							"--bind_ip_all",
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 27017}},
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(mongo, obj, r.Scheme); err != nil {
		return err
	}

	var existing appsv1.Deployment
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

func joinHosts(hosts []string) string {
	return strings.Join(hosts, ",")
}
