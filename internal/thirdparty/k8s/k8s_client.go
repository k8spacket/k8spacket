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
	"log/slog"
	"os"
	"slices"
	"strconv"
	"sync"
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
	Client
}

type SafeMap struct {
	mu   sync.RWMutex
	data map[string]ipResourceInfo
}

var k8sInfo *SafeMap

var clientset *kubernetes.Clientset

var disabledK8sResource, _ = strconv.ParseBool(os.Getenv("K8S_PACKET_K8S_RESOURCES_DISABLED"))

func init() {
	k8sInfo = &SafeMap{data: make(map[string]ipResourceInfo)}
	if !disabledK8sResource {
		_, clientset = configClusterClient()
		factory := informers.NewSharedInformerFactoryWithOptions(clientset, 5*time.Minute)
		stopChan := make(chan struct{})
		createPodInformer(factory)
		createSvcInformer(factory)
		createNodeInformer(factory)
		factory.Start(stopChan)
	}
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
}

func addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	ipResourceInfo := ipResourceInfo{
		ipResourceInfoType: Pod,
		Name:               "pod." + pod.Name,
		Namespace:          pod.Namespace,
	}
	addItem(pod.Status.PodIP, ipResourceInfo)
	slog.Debug("Added pod", "Name", pod.Name, "Namespace", pod.Namespace, "IP", pod.Status.PodIP)
}

func createSvcInformer(factory informers.SharedInformerFactory) {
	svcInformer := factory.Core().V1().Services().Informer()

	svcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addSvc(obj)
		},
		UpdateFunc: func(oldObj interface{}, obj interface{}) {
			addSvc(obj)
		},
	})
}

func addSvc(obj interface{}) {
	svc := obj.(*v1.Service)
	ipResourceInfo := ipResourceInfo{
		ipResourceInfoType: Svc,
		Name:               "svc." + svc.Name,
		Namespace:          svc.Namespace,
	}
	addItem(svc.Spec.ClusterIP, ipResourceInfo)
	slog.Debug("Added svc", "Name", svc.Name, "Namespace", svc.Namespace, "IP", svc.Spec.ClusterIP)
}

func createNodeInformer(factory informers.SharedInformerFactory) {
	nodeInformer := factory.Core().V1().Nodes().Informer()

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addNode(obj)
		},
		UpdateFunc: func(oldObj interface{}, obj interface{}) {
			addNode(obj)
		}})
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
			slog.Debug("Added node", "Name", node.Name, "IP", address.Address)
			break
		}
	}
}

func GetNameAndNamespace(id string) (string, string) {
	k8sInfo.mu.RLock()
	defer k8sInfo.mu.RUnlock()
	item, ok := k8sInfo.data[id]
	if ok {
		return item.Name, item.Namespace
	} else {
		return "", ""
	}
}

func addItem(id string, info ipResourceInfo) {
	k8sInfo.mu.Lock()
	defer k8sInfo.mu.Unlock()
	if k8sInfo.data[id].ipResourceInfoType != Node {
		k8sInfo.data[id] = info
	}
}
