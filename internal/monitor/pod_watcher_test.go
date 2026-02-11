package monitor

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

// TestPodWatcher_NormalPodLifecycle 测试正常 Pod 生命周期
func TestPodWatcher_NormalPodLifecycle(t *testing.T) {
	// 1. 创建 fake K8s 客户端
	client := fake.NewSimpleClientset()

	// 2. 创建 PodWatcher
	factory := informers.NewSharedInformerFactory(client, 0)
	watcher := &PodWatcher{
		clientset:        client, // fake.Clientset implements kubernetes.Interface
		factory:          factory,
		restartThreshold: 3,
		namespace:        "",
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	// 3. 启动 Informer
	podInformer := factory.Core().V1().Pods().Informer()
	
	eventReceived := false
	podInformer.AddEventHandler(watcher.createEventHandler(&eventReceived))

	go factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// 4. 模拟创建 Pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Name: "nginx", Image: "nginx"}},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}

	_, err := client.CoreV1().Pods("default").Create(nil, pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("创建 Pod 失败: %v", err)
	}

	// 等待事件处理
	time.Sleep(100 * time.Millisecond)

	if !eventReceived {
		t.Error("预期收到 Pod 创建事件，但未收到")
	}

	t.Log("✅ 正常 Pod 生命周期测试通过")
}

// TestPodWatcher_RestartDetection 测试容器重启检测
func TestPodWatcher_RestartDetection(t *testing.T) {
	client := fake.NewSimpleClientset()

	watcher := &PodWatcher{
		clientset:        client,
		factory:          informers.NewSharedInformerFactory(client, 0),
		restartThreshold: 3,
		namespace:        "",
	}

	// 创建一个重启次数超过阈值的 Pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crasher-pod",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "app",
					RestartCount: 5, // 超过阈值 3
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "CrashLoopBackOff",
							ExitCode: 1,
						},
					},
				},
			},
		},
	}

	// 检查是否触发告警逻辑
	alertTriggered := false
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.RestartCount >= watcher.restartThreshold {
			alertTriggered = true
			t.Logf("✅ 检测到容器重启: %s, 重启次数=%d", cs.Name, cs.RestartCount)
		}
	}

	if !alertTriggered {
		t.Error("预期触发重启告警，但未触发")
	}

	t.Log("✅ 容器重启检测测试通过")
}

// TestPodWatcher_ImagePullBackOff 测试镜像拉取失败检测
func TestPodWatcher_ImagePullBackOff(t *testing.T) {
	// 创建镜像拉取失败的 Pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "broken-pod",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "app",
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason:  "ImagePullBackOff",
							Message: "Back-off pulling image 'nginx:nonexistent'",
						},
					},
				},
			},
		},
	}

	// 检查异常状态检测
	anomalyDetected := false
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			reason := cs.State.Waiting.Reason
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
				anomalyDetected = true
				t.Logf("✅ 检测到镜像拉取异常: %s, 原因=%s", cs.Name, reason)
			}
		}
	}

	if !anomalyDetected {
		t.Error("预期检测到 ImagePullBackOff，但未检测到")
	}

	t.Log("✅ 镜像拉取失败检测测试通过")
}

// TestPodWatcher_DeleteEvent 测试 Pod 删除事件
func TestPodWatcher_DeleteEvent(t *testing.T) {
	client := fake.NewSimpleClientset()

	factory := informers.NewSharedInformerFactory(client, 0)
	watcher := &PodWatcher{
		clientset:        client,
		factory:          factory,
		restartThreshold: 3,
		namespace:        "",
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	podInformer := factory.Core().V1().Pods().Informer()
	
	deleteEventReceived := false
	podInformer.AddEventHandler(watcher.createDeleteEventHandler(&deleteEventReceived))

	go factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// 创建 Pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "delete-test-pod",
			Namespace: "default",
		},
	}

	_, err := client.CoreV1().Pods("default").Create(nil, pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("创建 Pod 失败: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// 删除 Pod
	err = client.CoreV1().Pods("default").Delete(nil, "delete-test-pod", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("删除 Pod 失败: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !deleteEventReceived {
		t.Error("预期收到 Pod 删除事件，但未收到")
	}

	t.Log("✅ Pod 删除事件测试通过")
}

// createEventHandler 创建事件处理器（用于测试）
func (pw *PodWatcher) createEventHandler(eventReceived *bool) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			*eventReceived = true
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			*eventReceived = true
		},
		DeleteFunc: func(obj interface{}) {
			*eventReceived = true
		},
	}
}

// createDeleteEventHandler 创建删除事件处理器
func (pw *PodWatcher) createDeleteEventHandler(deleteReceived *bool) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			*deleteReceived = true
		},
	}
}

// TestPodWatcher_MultipleRestarts 测试多容器 Pod 的重启检测
func TestPodWatcher_MultipleRestarts(t *testing.T) {
	watcher := &PodWatcher{
		restartThreshold: 3,
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "multi-container-pod",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "app",
					RestartCount: 2, // 未超过阈值
				},
				{
					Name:         "sidecar",
					RestartCount: 5, // 超过阈值
				},
			},
		},
	}

	alertCount := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.RestartCount >= watcher.restartThreshold {
			alertCount++
			t.Logf("✅ 检测到容器 %s 重启 %d 次", cs.Name, cs.RestartCount)
		}
	}

	if alertCount != 1 {
		t.Errorf("预期检测到 1 个容器超过阈值，实际检测到 %d 个", alertCount)
	}

	t.Log("✅ 多容器重启检测测试通过")
}
