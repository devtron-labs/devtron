package configMap

import (
	v1 "k8s.io/api/core/v1"
)

type Option func(*v1.ConfigMap)

// WithCmName adds config map name to a ConfigMap
func WithCmName(name string) Option {
	return func(cm *v1.ConfigMap) {
		if len(name) > 0 {
			cm.ObjectMeta.Name = name
		}
	}
}

// WithLabels adds labels to a ConfigMap
func WithLabels(labels map[string]string) Option {
	return func(cm *v1.ConfigMap) {
		if labels != nil && len(labels) > 0 {
			cm.ObjectMeta.Labels = labels
		}
	}
}

// WithAnnotations adds annotations to a ConfigMap
func WithAnnotations(annotations map[string]string) Option {
	return func(cm *v1.ConfigMap) {
		if annotations != nil && len(annotations) > 0 {
			cm.ObjectMeta.Annotations = annotations
		}
	}
}

// WithData adds string data to a ConfigMap
func WithData(data map[string]string) Option {
	return func(cm *v1.ConfigMap) {
		if data != nil && len(data) > 0 {
			cm.Data = data
		}
	}
}

// WithBinaryData adds binary data to a ConfigMap
func WithBinaryData(binaryData map[string][]byte) Option {
	return func(cm *v1.ConfigMap) {
		if binaryData != nil && len(binaryData) > 0 {
			cm.BinaryData = binaryData
		}
	}
}
