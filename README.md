pod-status-logger
=================

## Scope

This service watches k8s for pod events and logs them as JSON.

## Why?

We need a way to monitor/log/trace the lifecycle of pods and correlate with bundles.
The JSON logs can then be feeded into loghouse where they can be used for debugging etc.

## What?

The service registers for pod events and constructs log messages with the following fields:

* all pod labels
* "pod": "the_pod_name"
* "action": {create|update|delete}
* when "action" == "update"
  * "state": {waiting|running|error|terminated}
  * when "state" == {waiting|error}
    * "reason": "the reason as reported by k8s"
    * "info": "the info as reported by k8s message field"

## Known problems

* currently only the first container of a pod is inspected.

## Example loghouse output:
```
2018-03-16 14:11:43.153723876 action: create level: info pod: failer-64d8cc85cd-th2g8
2018-03-16 14:20:11.107242503 action: update info: level: info pod: failer-64d8cc85cd-th2g8 reason: state: terminated
2018-03-16 14:20:08.265428079 action: update info: level: info pod: failer-64d8cc85cd-th2g8 reason: state: terminated
2018-03-16 14:20:08.206095367 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:18:15.683580865 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:18:04.667848394 action: update info: rpc error: code = Unknown desc = Error response from daemon: pull access denied for trusch/hello-worldd, repository does not exist or may require 'docker login' level: error pod: failer-64d8cc85cd-th2g8 reason: ErrImagePull state: error
2018-03-16 14:15:24.669367959 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:15:12.667559701 action: update info: rpc error: code = Unknown desc = Error response from daemon: pull access denied for trusch/hello-worldd, repository does not exist or may require 'docker login' level: error pod: failer-64d8cc85cd-th2g8 reason: ErrImagePull state: error
2018-03-16 14:13:48.667614651 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:13:35.698289774 action: update info: rpc error: code = Unknown desc = Error response from daemon: pull access denied for trusch/hello-worldd, repository does not exist or may require 'docker login' level: error pod: failer-64d8cc85cd-th2g8 reason: ErrImagePull state: error
2018-03-16 14:13:04.666028492 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:12:49.674769319 action: update info: rpc error: code = Unknown desc = Error response from daemon: pull access denied for trusch/hello-worldd, repository does not exist or may require 'docker login' level: error pod: failer-64d8cc85cd-th2g8 reason: ErrImagePull state: error
2018-03-16 14:12:32.668322186 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:12:18.670012557 action: update info: rpc error: code = Unknown desc = Error response from daemon: pull access denied for trusch/hello-worldd, repository does not exist or may require 'docker login' level: error pod: failer-64d8cc85cd-th2g8 reason: ErrImagePull state: error
2018-03-16 14:12:01.668016372 action: update info: Back-off pulling image "trusch/hello-worldd" level: warning pod: failer-64d8cc85cd-th2g8 reason: ImagePullBackOff state: waiting
2018-03-16 14:11:46.584172297 action: update info: rpc error: code = Unknown desc = Error response from daemon: pull access denied for trusch/hello-worldd, repository does not exist or may require 'docker login' level: error pod: failer-64d8cc85cd-th2g8 reason: ErrImagePull state: error
2018-03-16 14:11:43.225598801 action: update info: level: warning pod: failer-64d8cc85cd-th2g8 reason: ContainerCreating state: waiting
2018-03-16 14:11:43.200667075 action: update level: warning pod: failer-64d8cc85cd-th2g8
2018-03-16 14:20:11.146785088 action: delete level: info pod: failer-64d8cc85cd-th2g8
2018-03-16 14:20:11.107242503 action: update info: level: info pod: failer-64d8cc85cd-th2g8 reason: state: terminated
```
