package k8sclient

type IK8SClient interface {
	GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string
}