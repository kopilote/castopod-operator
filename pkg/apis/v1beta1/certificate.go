package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeCertificateReady = "CertificateReady"
)

func SetCertificateReady(object Object, msg ...string) {
	SetCondition(object, ConditionTypeCertificateReady, metav1.ConditionTrue, msg...)
}

func SetCertificateError(object Object, msg ...string) {
	SetCondition(object, ConditionTypeCertificateReady, metav1.ConditionFalse, msg...)
}
