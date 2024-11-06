package k8sclient

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"slices"
	"strconv"
	"time"
)

type ipResourceInfoType string

const (
	Node ipResourceInfoType = "Node"
	Pod  ipResourceInfoType = "Pod"
	Svc  ipResourceInfoType = "Svc"
)

type ipResourceInfo struct {
	ipResourceInfoType ipResourceInfoType
	Name               string
	Namespace          string
}

type K8SClient struct {
	IK8SClient
}

var k8sInfo = make(map[string]ipResourceInfo)

var clientset *kubernetes.Clientset

var disabledK8sResource, _ = strconv.ParseBool(os.Getenv("K8S_PACKET_K8S_RESOURCES_DISABLED"))

func Init() {
	_, clientset = configClusterClient()
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 5*time.Minute)
	go createPodInformer(factory)
	go createSvcInformer(factory)
	go createNodeInformer(factory)
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

func createPodInformer(factory informers.SharedInformerFactory) {
	podInformer := factory.Core().V1().Pods().Informer()
	podChan := make(chan struct{})
	defer close(podChan)

	_, err := podInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod := obj.(*v1.Pod)
			return slices.ContainsFunc(pod.Status.Conditions, func(condition v1.PodCondition) bool {
				return condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue
			})
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				addPod(obj)
			},
			UpdateFunc: func(oldObj interface{}, obj interface{}) {
				addPod(obj)
			},
		}})
	if err != nil {
		fmt.Println(err)
		return
	}

	podInformer.Run(podChan)
}

func addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	ipResourceInfo := ipResourceInfo{
		ipResourceInfoType: Pod,
		Name:               "pod." + pod.Name,
		Namespace:          pod.Namespace,
	}
	addItem(pod.Status.PodIP, ipResourceInfo)
	fmt.Printf("Added pod - Name: %s, Namespace - %v, IP - %s\n", pod.Name, pod.Namespace, pod.Status.PodIP)
}

func createSvcInformer(factory informers.SharedInformerFactory) {
	svcInformer := factory.Core().V1().Services().Informer()
	svcChan := make(chan struct{})
	defer close(svcChan)

	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addSvc(obj)
		},
		UpdateFunc: func(oldObj interface{}, obj interface{}) {
			addSvc(obj)
		},
	})

	svcInformer.Run(svcChan)
}

func addSvc(obj interface{}) {
	svc := obj.(*v1.Service)
	ipResourceInfo := ipResourceInfo{
		ipResourceInfoType: Svc,
		Name:               "svc." + svc.Name,
		Namespace:          svc.Namespace,
	}
	addItem(svc.Spec.ClusterIP, ipResourceInfo)
	fmt.Printf("Added svc - Name: %s, Namespace - %v, IP - %s\n", svc.Name, svc.Namespace, svc.Spec.ClusterIP)
}

func createNodeInformer(factory informers.SharedInformerFactory) {
	nodeInformer := factory.Core().V1().Nodes().Informer()
	nodeChan := make(chan struct{})
	defer close(nodeChan)

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addNode(obj)
		},
		UpdateFunc: func(oldObj interface{}, obj interface{}) {
			addNode(obj)
		}})

	nodeInformer.Run(nodeChan)
}

func addNode(obj interface{}) {
	node := obj.(*v1.Node)
	ipResourceInfo := ipResourceInfo{
		ipResourceInfoType: Node,
		Name:               "node." + node.Name,
		Namespace:          "N/A",
	}
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			addItem(address.Address, ipResourceInfo)
			fmt.Printf("Added node - Name: %s, IP - %s\n", node.Name, address.Address)
			break
		}
	}
}

func GetNameAndNamespace(id string) (string, string) {
	item := k8sInfo[id]
	return item.Name, item.Namespace
}

func addItem(id string, info ipResourceInfo) {
	if k8sInfo[id].ipResourceInfoType != Node {
		k8sInfo[id] = info
	}
}
