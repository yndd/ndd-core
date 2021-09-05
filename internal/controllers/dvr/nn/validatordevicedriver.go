/*
Copyright 2021 Wim Henderickx.

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
	"fmt"

	"github.com/pkg/errors"
	dvrv1 "github.com/yndd/ndd-core/apis/dvr/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Errors
	errFailedListDeviceDrivers = "failed to list device drivers"
)

func (v *NnValidator) ValidateDeviceDriver(ctx context.Context, namespace, name, kind string, port int) (c *corev1.Container, err error) {
	log := v.log.WithValues("namespace", namespace, "name", name, "kind", kind, "port", port)
	log.Debug("ValidateDeviceDriver")

	// environment parameters used in the device driver
	envNameSpace := corev1.EnvVar{
		Name: "POD_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	}
	envPodIP := corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	}

	// list device drivers to see if there is specific device driver information
	if namespace == "" {
		namespace = "default"
	}
	selectors := []client.ListOption{
		client.MatchingLabels{
			"ddriver-kind": kind,
		},
	}
	dds := &dvrv1.DeviceDriverList{}
	if err := v.client.List(ctx, dds, selectors...); err != nil {
		return nil, errors.Wrap(err, errFailedListDeviceDrivers)
	}
	for _, dd := range dds.Items {
		c = dd.Spec.Container
	}
	if c == nil {
		log.Debug("Using the default device driver configuration")
		// apply the default settings
		c = &corev1.Container{
			Name:            "nddriver-" + name,
			Image:           "yndd/ndd-gnmi:latest",
			ImagePullPolicy: corev1.PullAlways,
			//ImagePullPolicy: corev1.PullIfNotPresent,
			Args: []string{
				"start",
				"--grpc-server-address=" + ":" + fmt.Sprintf("%d", port),
				"--device-name=" + name,
				"--namespace=" + namespace,
				"--debug",
			},
			Env: []corev1.EnvVar{
				envNameSpace,
				envPodIP,
			},
			Command: []string{
				"/ddriver",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("20m"),
					corev1.ResourceMemory: resource.MustParse("32Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
		}
	} else {
		log.Debug("Using the specific device driver configuration")
		// update the argument/environment information, since this is specific for the container deployment
		c.Args = []string{
			"start",
			"--grpc-server-address=" + ":" + fmt.Sprintf("%d", port),
			"--device-name=" + name,
			"--namespace=" + namespace,
			"--debug",
		}
		c.Env = []corev1.EnvVar{
			envNameSpace,
			envPodIP,
		}
	}
	return c, nil
}
