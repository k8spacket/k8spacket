package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type IPResourceInfo struct {
	Name      string
	Namespace string
}

var K8sInfo = make(map[string]IPResourceInfo)

func FetchK8SInfo() {

	m := make(map[string]IPResourceInfo)

	err, clientset := configClusterClient()

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	for i := range pods.Items {
		ipResourceInfo := new(IPResourceInfo)
		pod := pods.Items[i]
		ipResourceInfo.Name = "pod." + pod.Name
		ipResourceInfo.Namespace = pod.Namespace
		m[pod.Status.PodIP] = *ipResourceInfo
	}

	services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	for i := range services.Items {
		ipResourceInfo := new(IPResourceInfo)
		service := services.Items[i]
		ipResourceInfo.Name = "svc." + service.Name
		ipResourceInfo.Namespace = service.Namespace
		m[service.Spec.ClusterIP] = *ipResourceInfo
	}
	addMap(K8sInfo, m)
}

func configClusterClient() (error, *kubernetes.Clientset) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	return err, clientset
}

func GetDaemonK8sPacketIps() []string {

	list := make([]string, 0)

	err, clientset := configClusterClient()

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	for i := range pods.Items {
		pod := pods.Items[i]
		if pod.Labels["name"] == os.Getenv("K8S_PACKET_NAME_LABEL_VALUE") {
			list = append(list, pod.Status.PodIP)
		}
	}

	return list
}

func addMap(a map[string]IPResourceInfo, b map[string]IPResourceInfo) {
	for k, v := range b {
		a[k] = v
	}
}
