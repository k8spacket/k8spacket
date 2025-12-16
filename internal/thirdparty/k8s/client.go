package k8sclient

type Client interface {
	GetPodIPsBySelectors(fieldSelector string, labelSelector string) []string
}
