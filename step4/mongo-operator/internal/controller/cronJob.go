package controller

import (
	"context"
	"fmt"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MongoDBClusterReconciler) reconcileBackupCronJob(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster) error {
	if !mongo.Spec.Backup.Enabled {
		return nil
	}

	schedule := mongo.Spec.Backup.Schedule
	mongoURI := fmt.Sprintf("mongodb://%s:27017", mongo.Name+"-mongos")
	s3Bucket := mongo.Spec.Backup.Bucket

	script := `
BACKUP_NAME="backup-$(date +%F-%H%M%S).gz"
mongodump --uri="` + mongoURI + `" --archive | gzip > $BACKUP_NAME
aws --endpoint-url=$AWS_ENDPOINT_URL s3 cp $BACKUP_NAME s3://` + s3Bucket + `/$BACKUP_NAME
`

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-backup", mongo.Name),
			Namespace: mongo.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:    "mongodump",
									Image:   "sadegh81/mongo-aws:latest",
									Command: []string{"sh", "-c"},
									Args:    []string{script},
									Env: []corev1.EnvVar{
										{
											Name: "AWS_ACCESS_KEY_ID",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: mongo.Spec.Backup.SecretRef.Name,
													},
													Key: "accessKey",
												},
											},
										},
										{
											Name: "AWS_SECRET_ACCESS_KEY",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: mongo.Spec.Backup.SecretRef.Name,
													},
													Key: "secretKey",
												},
											},
										},
										{
											Name:  "AWS_ENDPOINT_URL",
											Value: mongo.Spec.Backup.StorageEndpoint,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(mongo, cronJob, r.Scheme); err != nil {
		return err
	}

	var existing batchv1.CronJob
	err := r.Get(ctx, client.ObjectKey{Name: cronJob.Name, Namespace: cronJob.Namespace}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, cronJob)
		}
		return err
	}

	cronJob.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, cronJob)
}
