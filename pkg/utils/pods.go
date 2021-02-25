package utils

import v1 "k8s.io/api/core/v1"

// IsLogPod determines if collect log
func IsLogPod(pod *v1.Pod) bool {
	if pod.ObjectMeta.Annotations["log"] == "true"{
		return true
	}
	return false
}
