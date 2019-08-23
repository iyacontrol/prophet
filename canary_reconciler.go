package main

import (
	"context"

	"github.com/golang/glog"
	canaryv1 "github.com/iyacontrol/shareit/pkg/apis/canary/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var replicas int32 = 1

type canaryReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *canaryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	glog.Info("reconciling canary")

	ctx := context.Background()

	var cd canaryv1.Canary
	err := r.Get(ctx, req.NamespacedName, &cd)
	if errors.IsNotFound(err) {
		glog.Error("Could not find canary")
		return ctrl.Result{}, nil
	}

	if err != nil {
		glog.Errorf("unable to get canary: %v", err)
		return ctrl.Result{}, err
	}

	var deploy appsv1.Deployment
	err = r.Get(ctx, req.NamespacedName, &deploy)
	if err != nil {
		return ctrl.Result{}, err
	}

	canaryDeployName := req.Name + "-canary"

	canary := &appsv1.Deployment{
		TypeMeta: deploy.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        canaryDeployName,
			Namespace:   req.Namespace,
			Labels:      deploy.Labels,
			Annotations: deploy.Annotations,
		},
		Spec: deploy.Spec,
	}

	for containerName, image := range cd.Spec.Images {
		for i, c := range deploy.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				deploy.Spec.Template.Spec.Containers[i].Image = image
			}
		}
	}

	canary.Spec.Replicas = &replicas

	switch cd.Spec.Stage {
	case K8sDeployStageCanary:
		err = r.Create(ctx, canary)
		if err != nil {
			glog.Errorf("unable to create canary deployment of %s: %v", req.Name, err)
			return ctrl.Result{}, err
		}

		glog.Infof("create canary deployment name: %s", canaryDeployName)

	case K8sDeployStageRollBack:
		r.Delete(ctx, canary)
		if err != nil {
			glog.Errorf("unable to delete canary deployment of %s: %v", canaryDeployName, err)
			return ctrl.Result{}, err
		}

		glog.Infof("delete canary deployment name: %s", canaryDeployName)

	case K8sDeployStageRollup:
		r.Delete(ctx, canary)
		if err != nil {
			glog.Errorf("unable to delete canary deployment of %s: %v", canaryDeployName, err)
			return ctrl.Result{}, err
		}

		for containerName, image := range cd.Spec.Images {
			for i, c := range deploy.Spec.Template.Spec.Containers {
				if c.Name == containerName {
					deploy.Spec.Template.Spec.Containers[i].Image = image
				}
			}
		}

		err = r.Update(ctx, &deploy)
		if err != nil {
			glog.Errorf("unable to update deployment of %s: %v", req.Name, err)
			return ctrl.Result{}, err
		}

		glog.Infof("update container images : %v of  deployment name: %s", cd.Spec.Images, req.Name)

	default:
		glog.Errorf("cannot handle stage %v", cd.Spec.Stage)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
