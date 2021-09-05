/*
Copyright 2021 NDD.

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

package nn

import (
	"context"

	ndddvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type adder interface {
	Add(item interface{})
}

type EnqueueRequestForAllDeviceDriversWithRequests struct {
	client client.Client
}

// Create enqueues a request for all network nodes which pertains to the Device Driver.
func (e *EnqueueRequestForAllDeviceDriversWithRequests) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.Object, q)
}

// Update enqueues a request for all network nodes which pertains to the Device Driver.
func (e *EnqueueRequestForAllDeviceDriversWithRequests) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.ObjectOld, q)
	e.add(evt.ObjectNew, q)
}

// Delete enqueues a request for all network nodes which pertains to the Device Driver.
func (e *EnqueueRequestForAllDeviceDriversWithRequests) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.Object, q)
}

// Generic enqueues a request for all network nodes which pertains to the Device Driver.
func (e *EnqueueRequestForAllDeviceDriversWithRequests) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.add(evt.Object, q)
}

func (e *EnqueueRequestForAllDeviceDriversWithRequests) add(obj runtime.Object, queue adder) {
	dd, ok := obj.(*ndddvrv1.DeviceDriver)
	if !ok {
		return
	}

	nn := &ndddvrv1.NetworkNodeList{}
	if err := e.client.List(context.TODO(), nn); err != nil {
		return
	}

	for _, n := range nn.Items {
		// only enqueue if the network node device driver kind matches with the device driver label
		for k, v := range dd.GetLabels() {
			if k == "ddriver-kind" {
				if string(*n.Spec.DeviceDriverKind) == v {
					queue.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: n.GetName()}})
				}
			}
		}
	}
}
