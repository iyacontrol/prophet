package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/glog"
	cronhpav1 "github.com/iyacontrol/shareit/pkg/apis/cronhpa/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type cronhpaReconciler struct {
	client.Client
	scheme *runtime.Scheme

	//
	Image string
	Account string
}

func (r *cronhpaReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	glog.Info("reconciling cronhpa")

	ctx := context.Background()

	var ch cronhpav1.CronHpa
	cronHpaErr := r.Get(ctx, req.NamespacedName, &ch)
	// 其他错误则直接返回错误
	if cronHpaErr != nil && !errors.IsNotFound(cronHpaErr) {
		glog.Errorf("unable to get cronhpa: %v", cronHpaErr)
		return ctrl.Result{}, cronHpaErr
	}

	// delete reletaed cronjobs
	var jobs batchv1beta1.CronJobList
	err := r.List(ctx, &jobs, client.MatchingLabels(map[string]string{
		"app": ch.Spec.HpaName,
	}))
	if err != nil {
		glog.Errorf("unable to list cronjon: %v", err)
		return ctrl.Result{}, err
	}

	for _, job := range jobs.Items {
		err = r.Delete(ctx, &job)
		if err != nil {
			glog.Errorf("unable to delete cronjob: %v", err)
			return ctrl.Result{}, err
		}
	}

	if cronHpaErr != nil {
		glog.Infof("%s cronhpa has been deleted", ch.Name)
	} else {
		glog.Infof("%s cronhpa has been added or updated", ch.Name)
		// add cronjobs
		for _, cycle := range ch.Spec.Cycles {
			// handle
			capacity := strconv.Itoa(int(cycle.Capacity))

			cron := &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind:       CronJobkind,
					APIVersion: CronJobApiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", ch.Spec.HpaName, cycle.Hour),
					Namespace: ch.Namespace,
					Labels: map[string]string{
						"app": ch.Spec.HpaName,
					},
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule: fmt.Sprintf(ScheduleCronExp, cycle.Hour),
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template:                v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									Containers: []v1.Container{
										{
											Name: "cronhpa",
											Image: r.Image,
											ImagePullPolicy: v1.PullIfNotPresent,
											Env: []v1.EnvVar{
												{
													Name: EnvHpa,
													Value: ch.Spec.HpaName,
												},
												{
													Name: EnvCapacity,
													Value: capacity,
												},
											},
											Command: []string{
												"prophet",
											},
										},
									},
									ImagePullSecrets: []v1.LocalObjectReference{
										{
											Name: ImagePullSecret,
										},
									},
									ServiceAccountName: r.Account,
								},
							},
						},
					},
				},
			}

			err = r.Create(ctx, cron)
			if err != nil {
				glog.Errorf("unable to create cronjob: %v", err)
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}
