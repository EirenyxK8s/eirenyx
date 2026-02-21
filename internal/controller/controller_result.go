package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Complete() (ctrl.Result, error) {
	return ctrl.Result{Requeue: false}, nil
}

func CompleteWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{Requeue: false}, err
}

func Requeue(duration time.Duration) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true, RequeueAfter: duration}, nil
}
