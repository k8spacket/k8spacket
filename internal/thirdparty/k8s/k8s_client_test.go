package k8sclient

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNameAndNamespace_Empty(t *testing.T) {
	// Ensure k8s resources are disabled for init
	os.Setenv("K8S_PACKET_K8S_RESOURCES_DISABLED", "true")

	name, ns := GetNameAndNamespace("no-such-ip")
	assert.Equal(t, "", name)
	assert.Equal(t, "", ns)
}

func TestAddItem_NodeThenPodBehavior(t *testing.T) {
	os.Setenv("K8S_PACKET_K8S_RESOURCES_DISABLED", "true")

	// reset map
	k8sInfo = &SafeMap{data: make(map[string]ipResourceInfo)}

	// add Node first
	addItem("10.0.0.1", ipResourceInfo{ipResourceInfoType: Node, Name: "node.one", Namespace: "N/A"})
	name, ns := GetNameAndNamespace("10.0.0.1")
	assert.Equal(t, "node.one", name)
	assert.Equal(t, "N/A", ns)

	// attempt to add Pod for same IP - should NOT overwrite Node
	addItem("10.0.0.1", ipResourceInfo{ipResourceInfoType: Pod, Name: "pod.one", Namespace: "default"})
	name2, ns2 := GetNameAndNamespace("10.0.0.1")
	assert.Equal(t, "node.one", name2)
	assert.Equal(t, "N/A", ns2)
}

func TestAddItem_PodThenNodeBehavior(t *testing.T) {
	os.Setenv("K8S_PACKET_K8S_RESOURCES_DISABLED", "true")

	k8sInfo = &SafeMap{data: make(map[string]ipResourceInfo)}

	// add Pod first
	addItem("10.0.0.2", ipResourceInfo{ipResourceInfoType: Pod, Name: "pod.two", Namespace: "default"})
	name, ns := GetNameAndNamespace("10.0.0.2")
	assert.Equal(t, "pod.two", name)
	assert.Equal(t, "default", ns)

	// add Node for same IP - since existing is not Node, Node should overwrite
	addItem("10.0.0.2", ipResourceInfo{ipResourceInfoType: Node, Name: "node.two", Namespace: "N/A"})
	name2, ns2 := GetNameAndNamespace("10.0.0.2")
	assert.Equal(t, "node.two", name2)
	assert.Equal(t, "N/A", ns2)
}

func TestK8SClient_GetPodIPsBySelectors_Disabled(t *testing.T) {
	os.Setenv("K8S_PACKET_K8S_RESOURCES_DISABLED", "true")

	client := &K8SClient{}
	res := client.GetPodIPsBySelectors("", "")
	assert.Equal(t, []string{"127.0.0.1"}, res)
}
