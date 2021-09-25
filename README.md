![](https://github.com/supercaracal/aws-sqs-worker-job-controller/workflows/Test/badge.svg?branch=master)
![](https://github.com/supercaracal/aws-sqs-worker-job-controller/workflows/Release/badge.svg)

AWS SQS Worker Job Controller
=================================================

This is a custom controller for Kubernetes.  
The controller aims at handling worker jobs for queueing.  
A worker job is invoked by the controller with a message of the queue as args when successful dequeue.  
Worker jobs are declared by [CRD](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/). The custom resources own `Job` resources.  

```
Pod <- Job <- AWSSQSWorkerJob
```

## Running on local host
```
$ kind create cluster
$ make apply-manifests
$ make build
$ make run
```

## Running in Docker
```
$ kind create cluster
$ make apply-manifests
$ make build-image
$ make port-forward-registry &
$ make push-image
```

## Verify operation with LocalStack
```
$ make port-forward-localstack &
$ make create-example-queue
$ make enqueue-example-task
$ make get-example-queue-attrs
```

## See also
* [sample-controller](https://github.com/kubernetes/sample-controller)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [kind](https://github.com/kubernetes-sigs/kind)
* [Kubernetes Reference](https://kubernetes.io/docs/reference/)
* [(Japanese) Under the Kubernetes Controller](https://speakerdeck.com/govargo/under-the-kubernetes-controller-36f9b71b-9781-4846-9625-23c31da93014)
