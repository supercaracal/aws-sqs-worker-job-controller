package workers

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
	"k8s.io/klog/v2"

	"github.com/supercaracal/aws-sqs-worker-job-controller/internal/queues"
)

// Consumer is
type Consumer struct {
	queue                *queues.MessageQueue
	kubeClientSet        kubernetes.Interface
	customClientSet      clientset.Interface
	customResourceLister listers.AwsSqsWorkerJobLister
}

// NewConsumer is
func NewConsumer(
	region string,
	endpointURL string,
	kubeClientSet kubernetes.Interface,
	customClientSet clientset.Interface,
	customResourceLister listers.AwsSqsWorkerJobLister,
) (*Consumer, error) {
	q, err := queues.NewMySQS(region, endpointURL)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		queue:                q,
		kubeClientSet:        kubeClientSet,
		customClientSet:      customClientSet,
		customResourceLister: customResourceLister,
	}
}

// Run is
func (c *Consumer) Run() {
	jobs, err := getJobList(c.kubeClientSet)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Failed to extract job list: %v", err))
		return
	}
	klog.V(4).Infof("Found %d jobs", len(jobs))

	jobsByParent := groupJobsByParent(jobs)
	klog.V(4).Infof("Found %d groups", len(jobsByParent))
}

func getJobList(kubeClientSet kubernetes.Interface) ([]batchv1.Job, error) {
	jobListFunc := func(opts metav1.ListOptions) (runtime.Object, error) {
		return kubeClientSet.BatchV1().Jobs(metav1.NamespaceAll).List(context.TODO(), opts)
	}

	jobs := make([]batchv1.Job, 0)
	err := pager.
		New(pager.SimplePageFunc(jobListFunc)).
		EachListItem(context.Background(), metav1.ListOptions{}, func(obj runtime.Object) error {
			tmp, ok := obj.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("expected type *batchv1.Job, got type %T", tmp)
			}
			jobs = append(jobs, *tmp)
			return nil
		})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func groupJobsByParent(jobs []batchv1.Job) map[types.UID][]batchv1.Job {
	jobsByParent := make(map[types.UID][]batchv1.Job)
	for _, job := range js {
		parentUID, found := getParentUIDFromJob(job)
		if !found {
			klog.V(4).Infof("Unable to get parent uid from job %s in namespace %s", job.Name, job.Namespace)
			continue
		}
		jobsByParent[parentUID] = append(jobsByParent[parentUID], job)
	}
	return jobsByParent
}

func getParentUIDFromJob(j batchv1.Job) (types.UID, bool) {
	controllerRef := metav1.GetControllerOf(&j)

	if controllerRef == nil {
		return types.UID(""), false
	}

	if controllerRef.Kind != "AwsSqsWorkerJob" {
		klog.V(4).Infof("Job with non-AwsSqsWorkerJob parent, name %s namespace %s", j.Name, j.Namespace)
		return types.UID(""), false
	}

	return controllerRef.UID, true
}

func syncAll(customClientSet) error {
	customResourceListFunc := func(opts metav1.ListOptions) (runtime.Object, error) {
		return customClientSet.BatchV1beta1().CronJobs(metav1.NamespaceAll).List(context.TODO(), opts)
	}

	err = pager.New(pager.SimplePageFunc(cronJobListFunc)).EachListItem(context.Background(), metav1.ListOptions{}, func(object runtime.Object) error {
		cj, ok := object.(*batchv1beta1.CronJob)
		if !ok {
			return fmt.Errorf("expected type *batchv1beta1.CronJob, got type %T", cj)
		}
		syncOne(cj, jobsByCj[cj.UID], time.Now(), jm.jobControl, jm.cjControl, jm.recorder)
		cleanupFinishedJobs(cj, jobsByCj[cj.UID], jm.jobControl, jm.cjControl, jm.recorder)
		return nil
	})

	return err
}
