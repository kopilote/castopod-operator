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

package v1beta1

import (
	pkgv1beta1 "github.com/kopilote/castopod-operator/pkg/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigurationSpec defines the desired state of Configuration
type ConfigurationSpec struct {
	//+required
	Mysql MysqlConfig `json:"mysql"`
	//+required
	Ingress Ingress `json:"ingress,omitempty"`
	//+optional
	Media ConfigMedia `json:"media,omitempty"`
	//+optional
	Smtp ConfigSmtp `json:"smtp,omitempty"`
}

type ConfigSmtp struct {
	//+optional
	From     string `json:"from,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ConfigMedia struct {
	//+optional
	Endpoint string `json:"endpoint,omitempty"`
	Key      string `json:"key,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Region   string `json:"region,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Configuration is the Schema for the castopods API
type Configuration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigurationSpec `json:"spec,omitempty"`
	Status pkgv1beta1.Status `json:"status,omitempty"`
}

func (*Configuration) Hub() {}

//+kubebuilder:object:root=true

// ConfigurationList contains a list of Configuration
type ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Configuration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Configuration{}, &ConfigurationList{})
}
