// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build kubeapiserver

package mutate

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_contains(t *testing.T) {
	type args struct {
		envs []corev1.EnvVar
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "contains",
			args: args{
				envs: []corev1.EnvVar{
					{Name: "foo", Value: "bar"},
					{Name: "baz", Value: "bar"},
				},
				name: "baz",
			},
			want: true,
		},
		{
			name: "doesn't contain",
			args: args{
				envs: []corev1.EnvVar{
					{Name: "foo", Value: "bar"},
					{Name: "baz", Value: "bar"},
				},
				name: "baf",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.envs, tt.args.name); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_injectEnv(t *testing.T) {
	type args struct {
		pod *corev1.Pod
		env corev1.EnvVar
	}
	tests := []struct {
		name        string
		args        args
		wantPodFunc func() corev1.Pod
		wantErr     bool
	}{
		{
			name: "1 container, 1 inject env",
			args: args{
				pod: fakePodWithContainer("foo-pod", fakeContainer("foo-container")),
				env: fakeEnv("inject-me"),
			},
			wantPodFunc: func() corev1.Pod {
				pod := fakePodWithContainer("foo-pod", fakeContainer("foo-container"))
				pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, fakeEnv("inject-me"))
				return *pod
			},
		},
		{
			name: "1 container, 0 inject env",
			args: args{
				pod: fakePodWithContainer("foo-pod", fakeContainer("foo-container")),
				env: fakeEnv("foo-container-env-foo"),
			},
			wantPodFunc: func() corev1.Pod {
				return *fakePodWithContainer("foo-pod", fakeContainer("foo-container"))
			},
		},
		{
			name: "2 container, 2 inject env",
			args: args{
				pod: fakePodWithContainer("foo-pod", fakeContainer("foo-container"), fakeContainer("bar-container")),
				env: fakeEnv("inject-me"),
			},
			wantPodFunc: func() corev1.Pod {
				pod := fakePodWithContainer("foo-pod", fakeContainer("foo-container"), fakeContainer("bar-container"))
				pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, fakeEnv("inject-me"))
				pod.Spec.Containers[1].Env = append(pod.Spec.Containers[1].Env, fakeEnv("inject-me"))
				return *pod
			},
		},
		{
			name: "2 container, 1 inject env",
			args: args{
				pod: fakePodWithContainer("foo-pod", fakeContainer("foo-container"), fakeContainer("bar-container")),
				env: fakeEnv("foo-container-env-foo"),
			},
			wantPodFunc: func() corev1.Pod {
				pod := fakePodWithContainer("foo-pod", fakeContainer("foo-container"), fakeContainer("bar-container"))
				pod.Spec.Containers[1].Env = append(pod.Spec.Containers[1].Env, fakeEnv("foo-container-env-foo"))
				return *pod
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injectEnv(tt.args.pod, tt.args.env)
			if tt.args.pod != nil && !reflect.DeepEqual(tt.args.pod.Spec.Containers, tt.wantPodFunc().Spec.Containers) {
				t.Errorf("injectEnv() = %v, want %v", tt.args.pod.Spec.Containers, tt.wantPodFunc().Spec.Containers)
			}
		})
	}
}

func fakeEnv(name string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: name + "-env-value",
	}
}

func fakeContainer(name string) corev1.Container {
	return corev1.Container{
		Name: name,
		Env: []corev1.EnvVar{
			fakeEnv(name + "-env-foo"),
			fakeEnv(name + "-env-bar"),
		},
	}
}

func fakePodWithContainer(name string, containers ...corev1.Container) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	}
}
