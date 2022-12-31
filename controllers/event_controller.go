/*
Copyright 2022.

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
	"github.com/heimdall-controller/slack-prototype/controllers/slack"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type EventController struct {
	client.Client
	*runtime.Scheme
}

var _ reconcile.Reconciler = &EventController{}

// Add +kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch
func (ec EventController) Add(mgr manager.Manager, selector v12.LabelSelector) error {
	// Create a new Controller
	c, err := controller.New("event-controller", mgr,
		controller.Options{Reconciler: &EventController{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}})
	if err != nil {
		logrus.Errorf("failed to create pod controller: %v", err)
		return err
	}

	// Create label selector containing the specified label
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(selector)
	if err != nil {
		logrus.Errorf("Error creating label selector predicate: %v", err)
		return err
	}

	// Add a watch to objects containing that label
	err = c.Watch(
		&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForObject{}, labelSelectorPredicate)
	if err != nil {
		logrus.Errorf("Error creating watch for objects: %v", err)
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

func (ec EventController) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	logrus.Infof("reconciling pod %s", req.NamespacedName)

	pod := &v1.Pod{}
	err := ec.Client.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get secret
	slackSecret := &v1.Secret{}

	err = ec.Client.Get(ctx, client.ObjectKey{
		Name:      "slack-credentials",
		Namespace: "default",
	}, slackSecret)

	if err != nil {
		if errors.IsNotFound(err) {
			err := ec.Client.Create(ctx, slackSecret)
			if err != nil {
				return ctrl.Result{}, err
			}
			logrus.Infof("empty slack secret created: %s", slackSecret.Name)
		}
		logrus.Error(err, "failed to get slack credentials")
		return reconcile.Result{}, err
	}

	logrus.Infof("found slack secret: %s", slackSecret.Name)

	slack.SendEvent(pod, slackSecret)

	return ctrl.Result{}, nil
}
