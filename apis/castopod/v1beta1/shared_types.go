package v1beta1

import (
	"fmt"
	"strings"

	"github.com/kopilote/castopod-operator/pkg/typeutils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type Ingress struct {
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	TLS IngressTLS `json:"tls"`
}

func (t *IngressTLS) AsK8SIngressTLSSlice() []networkingv1.IngressTLS {
	if t == nil {
		return nil
	}
	return []networkingv1.IngressTLS{{
		//SecretName: t.SecretName,
	}}
}

type IngressTLS struct {
	// SecretName is the name of the secret used to terminate TLS traffic on
	// port 443. Field is left optional to allow TLS routing based on SNI
	// hostname alone. If the SNI host in a listener conflicts with the "Host"
	// header field used by an IngressRule, the SNI host is used for termination
	// and value of the Host header is used for routing.
	// +optional
	SecretName string `json:"secretName,omitempty" protobuf:"bytes,2,opt,name=secretName"`
}

func (c *ConfigSource) Env() *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		ConfigMapKeyRef: c.ConfigMapKeyRef,
		SecretKeyRef:    c.SecretKeyRef,
	}
}

func SelectRequiredConfigValueOrReference[VALUE interface {
	string |
		int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64
}, SRC interface {
	*ConfigSource | *corev1.EnvVarSource
}](key, prefix string, v VALUE, src SRC) corev1.EnvVar {
	var (
		ret         corev1.EnvVar
		stringValue *string
	)
	switch v := any(v).(type) {
	case string:
		if v != "" {
			stringValue = &v
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if v != 0 {
			value := fmt.Sprintf("%d", v)
			stringValue = &value
		}
	}
	if stringValue != nil {
		ret = typeutils.EnvWithPrefix(prefix, key, *stringValue)
	} else {
		switch src := any(src).(type) {
		case *ConfigSource:
			ret = typeutils.EnvFromWithPrefix(prefix, key, src.Env())
		case *corev1.EnvVarSource:
			ret = corev1.EnvVar{
				Name:      prefix + key,
				ValueFrom: src,
			}
		}
	}
	return ret
}

type MysqlConfig struct {
	// +optional
	Port int `json:"port"`
	// +optional
	PortFrom *ConfigSource `json:"portFrom"`
	// +optional
	Host string `json:"host"`
	// +optional
	HostFrom *ConfigSource `json:"hostFrom"`
	// +optional
	Username string `json:"username"`
	// +optional
	UsernameFrom *ConfigSource `json:"usernameFrom"`
	// +optional
	Password string `json:"password"`
	// +optional
	PasswordFrom *ConfigSource `json:"passwordFrom"`
}

func (c *MysqlConfig) EnvWithDiscriminator(prefix, discriminator string) []corev1.EnvVar {

	discriminator = strings.ToUpper(discriminator)
	withDiscriminator := func(v string) string {
		if discriminator == "" {
			return v
		}
		return fmt.Sprintf("%s_%s", v, discriminator)
	}

	ret := make([]corev1.EnvVar, 0)
	ret = append(ret, SelectRequiredConfigValueOrReference(withDiscriminator("CP_DATABASE_HOSTNAME"), prefix, c.Host, c.HostFrom))
	ret = append(ret, SelectRequiredConfigValueOrReference(withDiscriminator("CP_DATABASE_PORT"), prefix, c.Port, c.PortFrom))
	ret = append(ret, SelectRequiredConfigValueOrReference(withDiscriminator("database.default.port"), prefix, c.Port, c.PortFrom))
	ret = append(ret, SelectRequiredConfigValueOrReference(withDiscriminator("CP_DATABASE_USERNAME"), prefix, c.Username, c.UsernameFrom))
	ret = append(ret, SelectRequiredConfigValueOrReference(withDiscriminator("CP_DATABASE_PASSWORD"), prefix, c.Password, c.PasswordFrom))

	return ret
}

func (c *MysqlConfig) Env(prefix string) []corev1.EnvVar {
	return c.EnvWithDiscriminator(prefix, "")
}

type ConfigSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty" protobuf:"bytes,3,opt,name=configMapKeyRef"`
	// Selects a key of a secret in the pod's namespace
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,4,opt,name=secretKeyRef"`
}
