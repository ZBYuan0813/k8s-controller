package pkg

import (
	"fmt"
	"log"

	"github.com/zanghao2/k8s-controller/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	clientgocache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

var (
	KeyFunc = clientgocache.DeletionHandlingMetaNamespaceKeyFunc
)

// Controller struct
type Controller struct {
	clientSet *kubernetes.Clientset
	podLister corelisters.PodLister
	podQueue  workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController new controller
func NewController(clientset *kubernetes.Clientset, kubeInformerFactory kubeinformers.SharedInformerFactory, stopCh <-chan struct{}) (*Controller, error) {
	log.Printf("info: Creating event broadcaster")
	// init event record
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "controller"})

	// new controller
	c := &Controller{
		clientSet: clientset,
		podQueue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "podQueue"),
		recorder:  recorder,
	}

	// create pod informer
	podInformer := kubeInformerFactory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(clientgocache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *v1.Pod:
				return utils.IsLogPod(t)
			case clientgocache.DeletedFinalStateUnknown:
				if pod, ok := t.Obj.(*v1.Pod); ok {
					log.Printf("debug: delete pods %s in ns %s", pod.Name, pod.Namespace)
					return utils.IsLogPod(pod)
				}
				runtime.HandleError(fmt.Errorf("unable to handle object in %T: %T", c, obj))
				return false
			default:
				runtime.HandleError(fmt.Errorf("unable to handle object in %T: %T", c, obj))
				return false
			}
		},
		Handler: clientgocache.ResourceEventHandlerFuncs{
			AddFunc:    c.addPodToCache,
			UpdateFunc: c.updatePodInCache,
			DeleteFunc: c.deletePodFromCache,
		},
	})

	return new(Controller), nil
}

// Run controller
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	log.Println("runing controller")
	return nil
}

func (c *Controller) addPodToCache(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Printf("warn: cannot convert to *v1.Pod: %v", obj)
		return
	}

	podKey, err := KeyFunc(pod)
	if err != nil {
		log.Printf("warn: Failed to get the jobkey: %v", err)
		return
	}

	c.podQueue.Add(podKey)
}

func (c *Controller) updatePodInCache(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		log.Printf("warn: cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		log.Printf("warn: cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}

	if oldPod.Status.Phase == newPod.Status.Phase{
		return
	}

	return
}

func (c *Controller) deletePodFromCache(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case clientgocache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			log.Printf("warn: cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		log.Printf("warn: cannot convert to *v1.Pod: %v", t)
		return
	}

	log.Printf("debug: delete pod %s in ns %s", pod.Name, pod.Namespace)
	podKey, err := KeyFunc(pod)
	if err != nil {
		log.Printf("warn: Failed to get the jobkey: %v", err)
		return
	}
	c.podQueue.Add(podKey)
}
