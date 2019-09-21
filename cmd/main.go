// Copied from
// https://raw.githubusercontent.com/kubernetes/client-go/master/examples/out-of-cluster-client-configuration/main.go
// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap/
// and modified slightly

package main

import (
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
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

	//#region USE PLAIN WATCH
	// https://stackoverflow.com/questions/40975307/how-to-watch-events-on-a-kubernetes-service-using-its-go-client
	//watch, err := clientset.CoreV1().Pods("").Watch(metav1.ListOptions{})
	//
	//if err != nil {
	//	panic(err.Error())
	//}
	//
	//for event := range watch.ResultChan() {
	//	fmt.Printf("Event type: %v\n", event.Type)
	//	fmt.Printf("Event: %v\n", event)
	//	p, ok := event.Object.(*corev1.Pod)
	//	if !ok {
	//		log.Fatal("unexpected type")
	//	}
	//	fmt.Println(p.Status.ContainerStatuses)
	//	fmt.Println(p.Status.Phase)
	//}
	//#endregion

	//#region USE INFORMER
	// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap/
	// https://stackoverflow.com/questions/40975307/how-to-watch-events-on-a-kubernetes-service-using-its-go-client
	//watchList := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), string(v1.ResourcePods), v1.NamespaceAll, fields.Everything())
	//
	//// var store = cache.NewStore(nil)
	//
	//_, controller := cache.NewInformer(watchList, &v1.Pod{}, time.Second*30, cache.ResourceEventHandlerFuncs{
	//	AddFunc: func(obj interface{}) {
	//		fmt.Printf("pod added: %s \n", obj)
	//	},
	//	DeleteFunc: func(obj interface{}) {
	//		fmt.Printf("pod deleted: %s \n", obj)
	//
	//		// FOLLOWING DOESN'T WORK
	//		//item, exists, err := store.GetByKey("etcd-minikube")
	//		//if err != nil {
	//		//	fmt.Println("Error trying to get etcd pod from the watch store(cache)")
	//		//} else if !exists {
	//		//	fmt.Println("Cannot find etcd pod. Strange!")
	//		//} else {
	//		//	fmt.Printf("Etcd pod: %v", item)
	//		//}
	//	},
	//	UpdateFunc: func(oldObj, newObj interface{}) {
	//		fmt.Printf("pod updated: %s \n", oldObj)
	//	},
	//})
	//
	//stop := make(chan struct{})
	//defer close(stop)
	//go controller.Run(stop)
	//
	//for {
	//	time.Sleep(time.Second)
	//}
	//#endregion

	// region USE SHARED INFORMER
	// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap/
	// https://stackoverflow.com/questions/40975307/how-to-watch-events-on-a-kubernetes-service-using-its-go-client

	watchList := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), string(v1.ResourcePods), v1.NamespaceAll, fields.Everything())

	informer := cache.NewSharedInformer(watchList, &v1.Pod{}, time.Second*30)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Printf("pod added: %s \n", obj)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Printf("pod deleted: %s \n", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Printf("pod updated: %s \n", oldObj)
		},
	})

	stop := make(chan struct{})
	defer close(stop)
	go informer.Run(stop)

	for {
		time.Sleep(time.Second)
	}

	// endregion

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
