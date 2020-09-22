AWS SQS Worker Controller for Kubernetes
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
* [kustomize](https://github.com/kubernetes-sigs/kustomize)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md)
* [Example of how to create and manage Kubernetes Custom Resource Definition.](https://github.com/jinghzhu/KubernetesCRD)
* [(Japanese) Under the Kubernetes Controller](https://speakerdeck.com/govargo/under-the-kubernetes-controller-36f9b71b-9781-4846-9625-23c31da93014)
* [(Japanese) KubernetesのCustom Resource Definition(CRD)とCustom Controller](https://www.sambaiz.net/article/182/)
* [(Japanese) KubernetesのCRDまわりを整理する](https://qiita.com/cvusk/items/773e222e0971a5391a51)
