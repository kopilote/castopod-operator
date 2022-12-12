package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeNamespaceReady  = "NamespaceReady"
	ConditionTypeDeploymentReady = "DeploymentReady"
)

func SetNamespaceCreated(object Object, msg ...string) {
	SetCondition(object, ConditionTypeNamespaceReady, metav1.ConditionTrue, msg...)
}

func SetNamespaceError(object Object, msg ...string) {
	SetCondition(object, ConditionTypeNamespaceReady, metav1.ConditionFalse, msg...)
}

func SetDeploymentReady(object Object, msg ...string) {
	SetCondition(object, ConditionTypeDeploymentReady, metav1.ConditionTrue, msg...)
}

func SetDeploymentError(object Object, msg ...string) {
	SetCondition(object, ConditionTypeDeploymentReady, metav1.ConditionFalse, msg...)
}
