package k8s

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const DefaultKubeConf = "/root/.kube/config"
const WorkerLabelKey = "node-role.kubernetes.io/worker"

func GetKubeNodes(ctx context.Context, client *kubernetes.Clientset, onlyWorker, onlySchedulable bool) ([]string, error) {
	listOptions := metav1.ListOptions{}
	if onlyWorker {
		listOptions.LabelSelector = WorkerLabelKey
	}
	nodeList, err := client.CoreV1().Nodes().List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	var nodeSlice []string
	for _, node := range nodeList.Items {
		if onlySchedulable && node.Spec.Unschedulable {
			unschedulable := false
			for _, taint := range node.Spec.Taints {
				if taint.Effect == corev1.TaintEffectNoSchedule {
					unschedulable = true
					break
				}
			}
			if unschedulable {
				continue
			}
		}

		nodeSlice = append(nodeSlice, node.Name)
	}
	return nodeSlice, nil
}

func CopyConfigmap(ctx context.Context, client *kubernetes.Clientset, oriNamespace, namespace, name string) error {
	_, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			err = nil
		} else {
			return err
		}

		// not exist: get configmap from oriNamespace and create at namespace
		var cm *corev1.ConfigMap
		if cm, err = client.CoreV1().ConfigMaps(oriNamespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
			return err
		}

		// new configmap
		newCm := &corev1.ConfigMap{}
		newCm.Namespace = namespace
		newCm.Name = name
		newCm.Data = cm.Data
		newCm.BinaryData = cm.BinaryData
		_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, newCm, metav1.CreateOptions{})
	}

	// exist, return
	return err
}

// NewClient
// if kubeConfPath == "", create k8s client auth by ServiceAccount in RBAC (/var/run/secrets/kubernetes.io/serviceaccount)
// otherwise kube client auth by kubeConfig in kubeConfPaths
func NewClient(kubeConfPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
