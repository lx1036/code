/*

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
// +kubebuilder:docs-gen:collapse=Apache License

/*
我们非常简单地开始：我们导入`meta / v1` API组，通常由其自身公开，但包含所有共同的元数据Kubernetes种类。
*/
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
接下来，我们为Kind定义Spec和Status。 Kubernetes功能通过使所需状态（“规范”）与实际群集状态（其他对象）协调一致
状态”和外部状态，然后记录观察到的状态（“状态”）。因此，每个* functional *对象都包含规范和状态。 几种类型，例如
ConfigMap不遵循这种模式，因为它们不对所需状态进行编码，但大多数类型都可以。
*/
// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobSpec defines the desired state of CronJob
type CronJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// CronJobStatus defines the observed state of CronJob
type CronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

/*
接下来，我们定义对应于实际种类的类型“ CronJob”和“ CronJobList”。
“ CronJob”是我们的根类型，描述了“ CronJob”的种类。像所有Kubernetes对象一样，它包含
TypeMeta（描述API版本和种类），还包含ObjectMeta，用于保存内容如名称，名称空间和标签。

`CronJobList`只是多个`CronJob`的容器。这是批量操作中使用的种类像LIST。

通常，我们绝不会修改其中的任何一种-所有修改都以“规范”或“状态”进行。

那个小小的 `+ kubebuilder：object：root` 注释被称为标记。我们拭目以待稍有更多，但是知道它们充当了额外的元数据，
[controller-tools](https://github.com/kubernetes-sigs/controller-tools)（我们的代码和YAML生成器）的其他信息。
这个特殊的告诉“ object”生成器，这种类型代表一种。然后，“对象”生成器生成
[runtime.Object]（https://godoc.org/k8s.io/apimachinery/pkg/runtime#Object）接口，这是我们的标准
所有代表Kinds的类型都必须实现的接口。
*/

// +kubebuilder:object:root=true

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSpec   `json:"spec,omitempty"`
	Status CronJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}

/*
最后，我们将Go类型添加到API组。 这使我们可以添加该API组中的任何类型 [Scheme](https://godoc.org/k8s.io/apimachinery/pkg/runtime#Scheme)。
*/
func init() {
	SchemeBuilder.Register(&CronJob{}, &CronJobList{})
}
