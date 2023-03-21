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

// CastopodSpec defines the desired state of Castopod
type CastopodSpec struct {
	//+required
	ConfigurationSpec string `json:"configurationSpec,omitempty"`
	//+required
	VersionSpec string `json:"versionSpec,omitempty"`
	//+required
	Activated bool `json:"activated"`
	//+optional
	Config Config `json:"config,omitempty"`
}

type Config struct {
	//+optional
	URL     ConfigUrl     `json:"url,omitempty"`
	Limit   ConfigLimit   `json:"limit,omitempty"`
	Gateway ConfigGateway `json:"gateway,omitempty"`
	Media   ConfigMedia   `json:"media,omitempty"`
	Smtp    ConfigSmtp    `json:"smtp,omitempty"`
}

type ConfigSmtp struct {
	//+optional
	From     string `json:"from,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ConfigMedia struct {
	//+optional
	Endpoint string `json:"endpoint,omitempty"`
	Key      string `json:"key,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Region   string `json:"region,omitempty"`
}

type ConfigUrl struct {
	Base        string `json:"base,omitempty"`
	Media       string `json:"media,omitempty"`
	LegalNotice string `json:"legalNotice,omitempty"`
}

type ConfigLimit struct {
	Storage   int64 `json:"storage,omitempty"`
	Bandwidth int64 `json:"bandwidth,omitempty"`
}

type ConfigGateway struct {
	Admin   string `json:"admin,omitempty"`
	Auth    string `json:"auth,omitempty"`
	Install string `json:"install,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Activated",type=boolean,JSONPath=`.spec.activated`
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.config.url.base`
//+kubebuilder:printcolumn:name="Configuration",type=string,JSONPath=`.spec.configurationSpec`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.versionSpec`

// Castopod is the Schema for the castopods API
type Castopod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CastopodSpec      `json:"spec,omitempty"`
	Status pkgv1beta1.Status `json:"status,omitempty"`
}

func (in *Castopod) IsDirty(t pkgv1beta1.Object) bool {
	return false
}

func (in *Castopod) GetStatus() pkgv1beta1.Dirty {
	return &in.Status
}

func (in *Castopod) GetConditions() *pkgv1beta1.Conditions {
	return &in.Status.Conditions
}

func NewStack(name string, spec CastopodSpec) Castopod {
	return Castopod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: spec,
	}
}

//+kubebuilder:object:root=true

// CastopodList contains a list of Castopod
type CastopodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Castopod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Castopod{}, &CastopodList{})
}
