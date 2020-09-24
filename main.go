package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	controllers "github.com/supercaracal/aws-sqs-worker-job-controller/internal/controllers"
	clientset "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	informers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
)

func setupSignalHandler() <-chan struct{} {
	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	stopCh := setupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	customClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building custom clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	customInformerFactory := informers.NewSharedInformerFactory(customClient, time.Second*30)

	customController := controllers.NewCustomController(
		kubeClient,
		customClient,
		kubeInformerFactory.Batch().V1().Jobs(),
		customInformerFactory.Awssqsworkerjobcontroller().V1().AwsSqsWorkerJobs(),
	)

	kubeInformerFactory.Start(stopCh)
	customInformerFactory.Start(stopCh)

	if err = customController.Run(stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(
		&masterURL,
		"master",
		"",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.",
	)

	flag.StringVar(
		&kubeconfig,
		"kubeconfig",
		"",
		"Path to a kubeconfig. Only required if out-of-cluster.",
	)
}
