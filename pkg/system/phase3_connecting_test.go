package system

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPodTargetPortNumber(t *testing.T) {
	tests := []struct {
		name       string
		pod        corev1.Pod
		targetPort intstr.IntOrString
		expected   int32
		expectedOK bool
	}{
		{
			name:       "numeric targetPort",
			pod:        corev1.Pod{},
			targetPort: intstr.FromInt32(6443),
			expected:   6443,
			expectedOK: true,
		},
		{
			name: "named port on main container",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Ports: []corev1.ContainerPort{{
							Name:          "iam-https",
							ContainerPort: 13443,
						}},
					}},
				},
			},
			targetPort: intstr.FromString("iam-https"),
			expected:   13443,
			expectedOK: true,
		},
		{
			name: "named port not found",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Ports: []corev1.ContainerPort{{
							Name:          "s3-https",
							ContainerPort: 6443,
						}},
					}},
				},
			},
			targetPort: intstr.FromString("iam-https"),
			expectedOK: false,
		},
		{
			name: "named port on second container",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Ports: []corev1.ContainerPort{{Name: "s3-https", ContainerPort: 6443}}},
						{Ports: []corev1.ContainerPort{{Name: "iam-https", ContainerPort: 13443}}},
					},
				},
			},
			targetPort: intstr.FromString("iam-https"),
			expected:   13443,
			expectedOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := podTargetPortNumber(&tt.pod, tt.targetPort)
			if ok != tt.expectedOK {
				t.Errorf("podTargetPortNumber() ok = %v, expected %v", ok, tt.expectedOK)
			}
			if got != tt.expected {
				t.Errorf("podTargetPortNumber() = %d, expected %d", got, tt.expected)
			}
		})
	}
}
