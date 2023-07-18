package utils

import (
	"context"
	"fmt"
	"github.com/vmware/cloud-provider-for-cloud-director/pkg/testingsdk"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	defaultRetryInterval = 20 * time.Second
	defaultRetryTimeout  = 600 * time.Second
)

func WaitForDeploymentReady(ctx context.Context, tc *testingsdk.TestClient, nameSpace string, deployName string) error {
	err := waitForDeploymentReady(ctx, tc.Cs.(*kubernetes.Clientset), nameSpace, deployName)
	if err != nil {
		return fmt.Errorf("error querying Deployment [%s] status for cluster [%s(%s)]: [%v]", deployName, tc.ClusterName, tc.ClusterId, err)
	}
	return nil
}

func waitForDeploymentReady(ctx context.Context, k8sClient *kubernetes.Clientset, nameSpace string, deployName string) error {
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		options := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", deployName),
		}
		podList, err := k8sClient.CoreV1().Pods(nameSpace).List(ctx, options)
		if err != nil {
			if testingsdk.IsRetryableError(err) {
				return false, nil
			}
			return false, fmt.Errorf("unexpected error occurred while getting deployment [%s]", deployName)
		}
		podCount := len(podList.Items)

		ready := 0
		for _, pod := range (*podList).Items {
			if pod.Status.Phase == apiv1.PodRunning {
				ready++
			}
		}
		if ready < podCount {
			fmt.Printf("running pods: %v < %v\n", ready, podCount)
			return false, nil
		}
		return true, nil
	})
	return err
}
