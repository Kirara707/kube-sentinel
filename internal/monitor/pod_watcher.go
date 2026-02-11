package monitor

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/Kirara707/kube-sentinel/pkg/logger"
)

// PodWatcher 基于 Informer 的 Pod 事件监听器
// 面试核心知识点：List-Watch 机制
//
// 工作原理：
//   1. List: 首次启动时全量获取所有 Pod 数据，写入本地缓存（Local Store）
//   2. Watch: 通过 HTTP 长连接监听 API Server 的增量变化
//   3. 本地缓存: 后续查询直接读内存，不再请求 API Server
//   4. Resync: 定期全量同步，防止本地缓存与 etcd 数据不一致
type PodWatcher struct {
	clientset        kubernetes.Interface // 使用接口便于单元测试 mock
	factory          informers.SharedInformerFactory
	restartThreshold int32
	namespace        string
}

// NewPodWatcher 创建 Pod 监听器
func NewPodWatcher(clientset kubernetes.Interface, resyncInterval time.Duration, restartThreshold int32, namespace string) *PodWatcher {
	var factory informers.SharedInformerFactory

	if namespace != "" {
		// 监听指定命名空间
		factory = informers.NewSharedInformerFactoryWithOptions(
			clientset,
			resyncInterval,
			informers.WithNamespace(namespace),
		)
	} else {
		// 监听所有命名空间
		factory = informers.NewSharedInformerFactory(clientset, resyncInterval)
	}

	return &PodWatcher{
		clientset:        clientset,
		factory:          factory,
		restartThreshold: restartThreshold,
		namespace:        namespace,
	}
}

// Start 启动 Informer 事件监听
// stopCh: 用于优雅关闭的通道
func (pw *PodWatcher) Start(stopCh <-chan struct{}) error {
	podInformer := pw.factory.Core().V1().Pods().Informer()

	// 注册事件处理函数 —— 事件驱动模式
	_, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// ========== Pod 创建事件 ==========
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				return
			}
			pw.handleAdd(pod)
		},

		// ========== Pod 更新事件（核心：检测重启） ==========
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod, ok1 := oldObj.(*v1.Pod)
			newPod, ok2 := newObj.(*v1.Pod)
			if !ok1 || !ok2 {
				return
			}
			pw.handleUpdate(oldPod, newPod)
		},

		// ========== Pod 删除事件 ==========
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				// 处理 DeletedFinalStateUnknown（Informer 的边缘场景）
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					logger.Log.Errorf("无法解析被删除的对象")
					return
				}
				pod, ok = tombstone.Obj.(*v1.Pod)
				if !ok {
					return
				}
			}
			pw.handleDelete(pod)
		},
	})

	if err != nil {
		return fmt.Errorf("注册事件处理器失败: %w", err)
	}

	// 启动所有 Informer
	pw.factory.Start(stopCh)

	// 等待缓存同步完成 —— 很重要！
	// 面试题：为什么要等待 WaitForCacheSync？
	// 答：确保本地缓存已经从 API Server 拉取了完整的数据，
	//     否则可能会在缓存未就绪时做出错误判断。
	logger.Log.Info("等待 Informer 缓存同步...")
	if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
		return fmt.Errorf("Informer 缓存同步超时")
	}

	logger.Log.Info("✅ Kube-Sentinel Informer 已启动，正在监听集群 Pod 事件...")
	return nil
}

// handleAdd 处理 Pod 新增事件
func (pw *PodWatcher) handleAdd(pod *v1.Pod) {
	event := PodHealthEvent{
		EventType: EventAdd,
		PodName:   pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		NodeName:  pod.Spec.NodeName,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Pod %s/%s 被创建", pod.Namespace, pod.Name),
	}

	logger.Log.Infow("🟢 Pod 新增",
		"pod", event.PodName,
		"namespace", event.Namespace,
		"phase", event.Phase,
		"node", event.NodeName,
	)

	// 检查新创建的 Pod 是否已经有异常状态
	pw.checkContainerStatuses(pod)
}

// handleUpdate 处理 Pod 更新事件
func (pw *PodWatcher) handleUpdate(oldPod, newPod *v1.Pod) {
	// 过滤无关紧要的更新（ResourceVersion 变化但实际状态没变）
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}

	// Phase 变化时记录
	if oldPod.Status.Phase != newPod.Status.Phase {
		logger.Log.Infow("🔄 Pod 状态变更",
			"pod", newPod.Name,
			"namespace", newPod.Namespace,
			"oldPhase", string(oldPod.Status.Phase),
			"newPhase", string(newPod.Status.Phase),
		)
	}

	// 核心：检测容器重启
	pw.checkContainerStatuses(newPod)
}

// handleDelete 处理 Pod 删除事件
func (pw *PodWatcher) handleDelete(pod *v1.Pod) {
	logger.Log.Infow("🔴 Pod 已删除",
		"pod", pod.Name,
		"namespace", pod.Namespace,
		"phase", string(pod.Status.Phase),
		"node", pod.Spec.NodeName,
	)
}

// checkContainerStatuses 检查容器状态，触发重启告警
// 这是监控的核心逻辑
func (pw *PodWatcher) checkContainerStatuses(pod *v1.Pod) {
	for _, cs := range pod.Status.ContainerStatuses {
		// 检测重启次数超过阈值
		if cs.RestartCount >= pw.restartThreshold {
			alert := RestartAlert{
				PodName:       pod.Name,
				Namespace:     pod.Namespace,
				ContainerName: cs.Name,
				RestartCount:  cs.RestartCount,
				Timestamp:     time.Now(),
			}

			// 解析上次终止原因
			if cs.LastTerminationState.Terminated != nil {
				alert.LastState = cs.LastTerminationState.Terminated.Reason
				alert.ExitCode = cs.LastTerminationState.Terminated.ExitCode
			}

			logger.Log.Warnw("⚠️  容器重启告警",
				"pod", alert.PodName,
				"namespace", alert.Namespace,
				"container", alert.ContainerName,
				"restartCount", alert.RestartCount,
				"lastState", alert.LastState,
				"exitCode", alert.ExitCode,
			)

			// TODO: 接入微信机器人 Webhook
			// 当 RestartCount > threshold 时，发送 HTTP Post 到微信机器人
			// pw.sendWeChatAlert(alert)
		}

		// 检测容器等待状态（CrashLoopBackOff、ImagePullBackOff 等）
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			reason := cs.State.Waiting.Reason
			if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" ||
				reason == "ErrImagePull" || reason == "CreateContainerConfigError" {
				logger.Log.Warnw("⚠️  容器异常状态",
					"pod", pod.Name,
					"namespace", pod.Namespace,
					"container", cs.Name,
					"reason", reason,
					"message", cs.State.Waiting.Message,
				)
			}
		}
	}

	// 同时检查 InitContainer 状态
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			logger.Log.Warnw("⚠️  InitContainer 异常",
				"pod", pod.Name,
				"namespace", pod.Namespace,
				"initContainer", cs.Name,
				"reason", cs.State.Waiting.Reason,
			)
		}
	}
}
