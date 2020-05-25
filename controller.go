package main

import (
	"context"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconcileNodeLabels struct {
	client client.Client
	log    logr.Logger
	labels map[string]string
}

// make sure reconcileNodeLabels implement the Reconciler interface
var _ reconcile.Reconciler = &reconcileNodeLabels{}

func (r *reconcileNodeLabels) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// set up a convenient log object so we don't have to type request over and over again
	log := r.log.WithValues("request", request)

	node := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, node)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find Node")
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Error(err, "Could not fetch Node")
		return reconcile.Result{}, err
	}

	// Set the label
	if node.Labels == nil {
		node.Labels = map[string]string{}
	}

	// Remove old labels
	for k := range node.Labels {
		if matchesLabelPrefix(k) {
			delete(node.Labels, k)
		}
	}

	for k, v := range r.labels {
		node.Labels[k] = v
	}

	err = r.client.Update(context.TODO(), node)
	if err != nil {
		log.Error(err, "Could not write Node")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
