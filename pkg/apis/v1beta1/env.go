package v1beta1

import (
	"fmt"

	"github.com/kopilote/castopod-operator/pkg/typeutils"
	corev1 "k8s.io/api/core/v1"
)

func EnvFromWithPrefix(prefix, key string, value *corev1.EnvVarSource) corev1.EnvVar {
	return corev1.EnvVar{
		Name:      prefix + key,
		ValueFrom: value,
	}
}

func EnvFrom(key string, value *corev1.EnvVarSource) corev1.EnvVar {
	return corev1.EnvVar{
		Name:      key,
		ValueFrom: value,
	}
}

func EnvWithPrefix(prefix, key, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  prefix + key,
		Value: value,
	}
}

func Env(key, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  key,
		Value: value,
	}
}
func EnvVarPlaceholder(key, prefix string) string {
	return fmt.Sprintf("$(%s%s)", prefix, key)
}

func ComputeEnvVar(prefix, format string, keys ...string) string {
	return fmt.Sprintf(format,
		typeutils.Map(keys, func(key string) any {
			return EnvVarPlaceholder(key, prefix)
		})...,
	)
}
