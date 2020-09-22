AWS SQS WorkerJob Controller for Kubernetes
=================================================

```
$ kind create cluster
$ kubectl config set-cluster kind --server=https://127.0.0.1:40963
$ make build
$ ./aws-sqs-worker-job-controller -kubeconfig=$HOME/.kube/config
```

# See also
* [kind](https://github.com/kubernetes-sigs/kind)
* [kubectl](https://github.com/kubernetes/kubectl)
* [code-generator](https://github.com/kubernetes/code-generator)
* [client-go](https://github.com/kubernetes/client-go)
* [sample-controller](https://github.com/kubernetes/sample-controller)
* [job-controller](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go)
* [cronjob-controller](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/cronjob/cronjob_controller.go)
* [apps-types](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/apps/types.go)
* [batch-types](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/batch/types.go)
* [Hit an unsupported type...](https://github.com/kubernetes/gengo/blob/7794989d00002eae09b50e95c3a221245260a20e/examples/deepcopy-gen/generators/deepcopy.go#L843-L886)
* [Test a weird version/kind embedding format.](https://github.com/kubernetes/apimachinery/blob/714f1137f89bf0ec6d038cf852d7661a1b9c660a/pkg/runtime/testing/types.go#L127-L156)
* [Command deepcopy-gen](https://godoc.org/k8s.io/gengo/examples/deepcopy-gen)
* [Extend the Kubernetes API with CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
* [Programming Kubernetes](https://www.amazon.com/dp/B07VCPM5VQ/)
* [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md)
* [Example of how to create and manage Kubernetes Custom Resource Definition.](https://github.com/jinghzhu/KubernetesCRD)
* [(Japanese) Under the Kubernetes Controller](https://speakerdeck.com/govargo/under-the-kubernetes-controller-36f9b71b-9781-4846-9625-23c31da93014)
* [(Japanese) Programming Kubernetesを読んで学んだこと](https://go-vargo.hatenablog.com/entry/2019/08/05/201546)
* [(Japanese) KubernetesのCustom Resource Definition(CRD)とCustom Controller](https://www.sambaiz.net/article/182/)
* [(Japanese) KubernetesのCRDまわりを整理する](https://qiita.com/cvusk/items/773e222e0971a5391a51)

# See also
* [kustomize](https://github.com/kubernetes-sigs/kustomize)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
