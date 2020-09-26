AWS SQS Worker Job Controller
=================================================

![](https://github.com/supercaracal/aws-sqs-worker-job-controller/workflows/Test/badge.svg)
![](https://github.com/supercaracal/aws-sqs-worker-job-controller/workflows/Docker/badge.svg)

This is a custom controller for Kubernetes.
The controller aims at handling a custom resource as a worker job for queueing.

# Docker

```
$ docker pull ghcr.io/supercaracal/aws-sqs-worker-job-controller:latest
```

# Kubernetes

```
$ kind create cluster
$ kubectl cluster-info --context kind-kind
$ make apply-manifests
$ kubectl port-forward service/localstack-service 4566:4566
```

```
$ aws --endpoint-url=http://localhost:4566 --region=us-west-2 sqs create-queue --queue-name=sleep-queue
{
    "QueueUrl": "http://localhost:4566/000000000000/sleep-queue"
}
$ aws --endpoint-url=http://localhost:4566 --region=us-west-2 sqs send-message --queue-url=http://localhost:4566/000000000000/sleep-queue --message-body=30
$ aws --endpoint-url=http://localhost:4566 --region=us-west-2 sqs get-queue-attributes --queue-url=http://localhost:4566/000000000000/sleep-queue --attribute-names=ApproximateNumberOfMessages
```

# Development

```
$ go get -u golang.org/x/lint/golint k8s.io/code-generator/...
$ make build
$ make run
```

# See also
* [Extend the Kubernetes API with CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
* [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md)
* [Programming Kubernetes](https://www.amazon.com/dp/B07VCPM5VQ/)

# See also
* [kubernetes/sample-controller](https://github.com/kubernetes/sample-controller)
* [programming-kubernetes/cnat-client-go](https://github.com/programming-kubernetes/cnat/tree/master/cnat-client-go)
* [uswitch/sqs-autoscaler-controller](https://github.com/uswitch/sqs-autoscaler-controller)

# See also
* [client-go](https://github.com/kubernetes/client-go)
* [job-controller](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/job_controller.go)
* [cronjob-controller](https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/cronjob/cronjob_controller.go)
* [core-types](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/types.go)
* [apps-types](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/apps/types.go)
* [batch-types](https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/batch/types.go)
* [A nice script for import kubernetes module](https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-521493597)

# See also
* [code-generator](https://github.com/kubernetes/code-generator)
* [Command deepcopy-gen](https://godoc.org/k8s.io/gengo/examples/deepcopy-gen)
* [Hit an unsupported type...](https://github.com/kubernetes/gengo/blob/7794989d00002eae09b50e95c3a221245260a20e/examples/deepcopy-gen/generators/deepcopy.go#L843-L886)
* [Test a weird version/kind embedding format.](https://github.com/kubernetes/apimachinery/blob/714f1137f89bf0ec6d038cf852d7661a1b9c660a/pkg/runtime/testing/types.go#L127-L156)

# See also
* [Example of how to create and manage Kubernetes Custom Resource Definition.](https://github.com/jinghzhu/KubernetesCRD)
* [(Japanese) Under the Kubernetes Controller](https://speakerdeck.com/govargo/under-the-kubernetes-controller-36f9b71b-9781-4846-9625-23c31da93014)
* [(Japanese) Programming Kubernetesを読んで学んだこと](https://go-vargo.hatenablog.com/entry/2019/08/05/201546)
* [(Japanese) KubernetesのCustom Resource Definition(CRD)とCustom Controller](https://www.sambaiz.net/article/182/)
* [(Japanese) KubernetesのCRDまわりを整理する](https://qiita.com/cvusk/items/773e222e0971a5391a51)

# See also
* [kind](https://github.com/kubernetes-sigs/kind)
* [kubectl](https://github.com/kubernetes/kubectl)
* [kustomize](https://github.com/kubernetes-sigs/kustomize)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
