/*
MIT License

Copyright (c) 2018 Contiamo Gmbh

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
	fields["namespace"] = pod.ObjectMeta.Namespace
	fields["cluster"] = pod.ObjectMeta.ClusterName
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
