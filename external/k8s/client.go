package k8sclient

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"strconv"
)

type IPResourceInfo struct {
	Name      string
	Namespace string
}

type K8SClient struct {
	IK8SClient
}

var _, clientset = configClusterClient()

var disabledK8sResource, _ = strconv.ParseBool(os.Getenv("K8S_PACKET_K8S_RESOURCES_DISABLED"))

func FetchK8SInfo() map[string]IPResourceInfo {

	if disabledK8sResource {
		fmt.Println("Getting k8s resources is disabled")
		return map[string]IPResourceInfo{}
	}

	fmt.Println("Getting k8s resources")

	m := make(map[string]IPResourceInfo)

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	fmt.Printf("Found %d pods\n", len(pods.Items))
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
	fmt.Printf("Found %d services\n", len(services.Items))
	for i := range services.Items {
		ipResourceInfo := new(IPResourceInfo)
		service := services.Items[i]
		ipResourceInfo.Name = "svc." + service.Name
		ipResourceInfo.Namespace = service.Namespace
		m[service.Spec.ClusterIP] = *ipResourceInfo
	}
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	fmt.Printf("Found %d nodes\n", len(nodes.Items))
	for i := range nodes.Items {
		ipResourceInfo := new(IPResourceInfo)
		node := nodes.Items[i]
		ipResourceInfo.Name = "node." + node.Name
		ipResourceInfo.Namespace = "N/A"
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				m[address.Address] = *ipResourceInfo
				break
			}
		}
	}
	return m
}

func (k8sClient *K8SClient) GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string {

	if disabledK8sResource {
		return []string{"127.0.0.1"}
	}

	list := make([]string, 0)

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector, LabelSelector: labelSelector})
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		list = append(list, pod.Status.PodIP)
	}

	return list
}

func configClusterClient() (error, *kubernetes.Clientset) {

	if disabledK8sResource {
		return nil, nil
	}

	config, err := rest.InClusterConfig()

	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}
	return err, cs
}
