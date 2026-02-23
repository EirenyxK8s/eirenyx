package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Complete() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func CompleteWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func RequeueAfter(d time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: d}, nil
}

func RequeueWithError(err error, d time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: d}, err
}
