// Copied from
// https://raw.githubusercontent.com/kubernetes/client-go/master/examples/out-of-cluster-client-configuration/main.go
// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap/
// and modified slightly

package main

import (
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// example code that shows how to handle errors
	namespace := "default"
	pod := "aaaaaa"
	_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	fmt.Println(err)
	if errors.IsNotFound(err) {
		fmt.Printf("Cannot find pod %v in namespace %v\n", pod, namespace)
	} else if err != nil {
		fmt.Printf("WTF?\n %v", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	for _, node := range nodes.Items {
		fmt.Printf("Status for the node: %v %v\n", node.Name, node.Status.Phase)

	}

	/// USE PLAIN WATCH

	watch, err := clientset.CoreV1().Pods("").Watch(metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	for event := range watch.ResultChan() {
		fmt.Printf("Event type: %v\n", event.Type)
		fmt.Printf("Event: %v\n", event)
		p, ok := event.Object.(*corev1.Pod)
		if !ok {
			log.Fatal("unexpected type")
		}
		fmt.Println(p.Status.ContainerStatuses)
		fmt.Println(p.Status.Phase)
	}

	// USE INFORMER
	// TODO:

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
