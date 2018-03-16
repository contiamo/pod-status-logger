/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func watchPods(kbc *kubernetes.Clientset) {
	resyncPeriod := 30 * time.Minute
	si := informers.NewSharedInformerFactory(kbc, resyncPeriod)
	si.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    podCreated,
			DeleteFunc: podDeleted,
			UpdateFunc: podUpdated,
		},
	)
	si.Start(wait.NeverStop)
}

func getLogger(pod *v1.Pod) *logrus.Entry {
	fields := make(logrus.Fields)
	fields["pod"] = pod.ObjectMeta.Name
	for k, v := range pod.ObjectMeta.Labels {
		fields[k] = v
	}
	return logrus.WithFields(fields)
}

func podCreated(obj interface{}) {
	pod := obj.(*v1.Pod)
	getLogger(pod).WithField("action", "create").Info()
}

func podUpdated(old, new interface{}) {
	newPod := new.(*v1.Pod)
	logger := getLogger(newPod).WithField("action", "update")
	var doWarning, doError bool
	if len(newPod.Status.ContainerStatuses) > 0 {
		stateObj := newPod.Status.ContainerStatuses[0].State
		switch {
		case stateObj.Running != nil:
			logger = logger.WithField("state", "running")
		case stateObj.Waiting != nil:
			if strings.HasPrefix(stateObj.Waiting.Reason, "Err") {
				logger = logger.WithField("state", "error")
				logger = logger.WithField("reason", stateObj.Waiting.Reason)
				logger = logger.WithField("info", stateObj.Waiting.Message)
				doError = true
			} else {
				logger = logger.WithField("state", "waiting")
				logger = logger.WithField("reason", stateObj.Waiting.Reason)
				logger = logger.WithField("info", stateObj.Waiting.Message)
				doWarning = true
			}
		case stateObj.Terminated != nil:
			logger = logger.WithField("state", "terminated")
		}
	} else {
		logger = logger.WithField("state", "waiting")
	}
	if len(newPod.Status.ContainerStatuses) < 1 || doWarning {
		logger.Warn()
	} else if doError {
		logger.Error()
	} else {
		logger.Info()
	}
}

func podDeleted(obj interface{}) {
	pod := obj.(*v1.Pod)
	getLogger(pod).WithField("action", "delete").Info()
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "can not create in-cluster config"))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "can not create clientset"))
	}
	watchPods(clientset)
	select {}
}
