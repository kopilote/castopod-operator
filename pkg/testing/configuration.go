package testing

import (
	"github.com/google/uuid"
	"github.com/kopilote/castopod-operator/apis/castopod/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDumbVersions() *v1beta1.Version {
	return &v1beta1.Version{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
		Spec: v1beta1.VersionSpec{
			ImageTag: "latest",
		},
	}
}

func NewDumbConfiguration() *v1beta1.Configuration {
	return &v1beta1.Configuration{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewString(),
		},
		Spec: v1beta1.ConfigurationSpec{
			Ingress: v1beta1.Ingress{
				Annotations: map[string]string{},
			},
			Mysql: NewDumpMysqlConfig(),
		},
	}
}

func NewDumpMysqlConfig() v1beta1.MysqlConfig {
	return v1beta1.MysqlConfig{
		Port:     3306,
		Host:     "postgres",
		Username: "username",
		Password: "password",
	}
}
