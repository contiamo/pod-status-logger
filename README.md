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
* "action": {create|update|delete}
* "pod": "POD_NAME"
* when "action" == "update"
  * "state": {waiting|running|error|terminated}
  * when "state" == {waiting|error}
    * "reason": "the reason as reported by k8s"
    * "info": "the info as reported by k8s message field"

## Known problems

* currently only the first container of a pod is inspected.
