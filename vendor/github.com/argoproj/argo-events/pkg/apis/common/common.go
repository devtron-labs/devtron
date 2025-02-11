/*
Copyright 2018 BlackRock, Inc.

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

package common

import (
	corev1 "k8s.io/api/core/v1"
)

// EventSourceType is the type of event source
type EventSourceType string

// possible event source types
var (
	MinioEvent           EventSourceType = "minio"
	CalendarEvent        EventSourceType = "calendar"
	FileEvent            EventSourceType = "file"
	ResourceEvent        EventSourceType = "resource"
	WebhookEvent         EventSourceType = "webhook"
	AMQPEvent            EventSourceType = "amqp"
	KafkaEvent           EventSourceType = "kafka"
	MQTTEvent            EventSourceType = "mqtt"
	NATSEvent            EventSourceType = "nats"
	SNSEvent             EventSourceType = "sns"
	SQSEvent             EventSourceType = "sqs"
	PubSubEvent          EventSourceType = "pubsub"
	GithubEvent          EventSourceType = "github"
	GitlabEvent          EventSourceType = "gitlab"
	HDFSEvent            EventSourceType = "hdfs"
	SlackEvent           EventSourceType = "slack"
	StorageGridEvent     EventSourceType = "storagegrid"
	AzureEventsHub       EventSourceType = "azureEventsHub"
	StripeEvent          EventSourceType = "stripe"
	EmitterEvent         EventSourceType = "emitter"
	RedisEvent           EventSourceType = "redis"
	RedisStreamEvent     EventSourceType = "redisStream"
	NSQEvent             EventSourceType = "nsq"
	PulsarEvent          EventSourceType = "pulsar"
	GenericEvent         EventSourceType = "generic"
	BitbucketServerEvent EventSourceType = "bitbucketserver"
	BitbucketEvent       EventSourceType = "bitbucket"
)

var (
	// RecreateStrategyEventSources refers to the list of event source types
	// that need to use Recreate strategy for its Deployment
	RecreateStrategyEventSources = []EventSourceType{
		AMQPEvent,
		CalendarEvent,
		KafkaEvent,
		PubSubEvent,
		AzureEventsHub,
		NATSEvent,
		MQTTEvent,
		MinioEvent,
		EmitterEvent,
		NSQEvent,
		PulsarEvent,
		RedisEvent,
		RedisStreamEvent,
		ResourceEvent,
		HDFSEvent,
		FileEvent,
		GenericEvent,
	}
)

// TriggerType is the type of trigger
type TriggerType string

// possible trigger types
var (
	OpenWhiskTrigger      TriggerType = "OpenWhisk"
	ArgoWorkflowTrigger   TriggerType = "ArgoWorkflow"
	LambdaTrigger         TriggerType = "Lambda"
	CustomTrigger         TriggerType = "Custom"
	HTTPTrigger           TriggerType = "HTTP"
	KafkaTrigger          TriggerType = "Kafka"
	PulsarTrigger         TriggerType = "Pulsar"
	LogTrigger            TriggerType = "Log"
	NATSTrigger           TriggerType = "NATS"
	SlackTrigger          TriggerType = "Slack"
	K8sTrigger            TriggerType = "Kubernetes"
	AzureEventHubsTrigger TriggerType = "AzureEventHubs"
)

// EventBusType is the type of event bus
type EventBusType string

// possible event bus types
var (
	EventBusNATS      EventBusType = "nats"
	EventBusJetStream EventBusType = "jetstream"
)

// BasicAuth contains the reference to K8s secrets that holds the username and password
type BasicAuth struct {
	// Username refers to the Kubernetes secret that holds the username required for basic auth.
	Username *corev1.SecretKeySelector `json:"username,omitempty" protobuf:"bytes,1,opt,name=username"`
	// Password refers to the Kubernetes secret that holds the password required for basic auth.
	Password *corev1.SecretKeySelector `json:"password,omitempty" protobuf:"bytes,2,opt,name=password"`
}

// SecureHeader refers to HTTP Headers with auth tokens as values
type SecureHeader struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Values can be read from either secrets or configmaps
	ValueFrom *ValueFromSource `json:"valueFrom,omitempty" protobuf:"bytes,2,opt,name=valueFrom"`
}

// ValueFromSource allows you to reference keys from either a Configmap or Secret
type ValueFromSource struct {
	SecretKeyRef    *corev1.SecretKeySelector    `json:"secretKeyRef,omitempty" protobuf:"bytes,1,opt,name=secretKeyRef"`
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty" protobuf:"bytes,2,opt,name=configMapKeyRef"`
}

// TLSConfig refers to TLS configuration for a client.
type TLSConfig struct {
	// CACertSecret refers to the secret that contains the CA cert
	CACertSecret *corev1.SecretKeySelector `json:"caCertSecret,omitempty" protobuf:"bytes,1,opt,name=caCertSecret"`
	// ClientCertSecret refers to the secret that contains the client cert
	ClientCertSecret *corev1.SecretKeySelector `json:"clientCertSecret,omitempty" protobuf:"bytes,2,opt,name=clientCertSecret"`
	// ClientKeySecret refers to the secret that contains the client key
	ClientKeySecret *corev1.SecretKeySelector `json:"clientKeySecret,omitempty" protobuf:"bytes,3,opt,name=clientKeySecret"`
	// If true, skips creation of TLSConfig with certs and creates an empty TLSConfig. (Defaults to false)
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty" protobuf:"varint,4,opt,name=insecureSkipVerify"`
}

// SASLConfig refers to SASL configuration for a client
type SASLConfig struct {
	// SASLMechanism is the name of the enabled SASL mechanism.
	// Possible values: OAUTHBEARER, PLAIN (defaults to PLAIN).
	// +optional
	Mechanism string `json:"mechanism,omitempty" protobuf:"bytes,1,opt,name=mechanism"`
	// User is the authentication identity (authcid) to present for
	// SASL/PLAIN or SASL/SCRAM authentication
	UserSecret *corev1.SecretKeySelector `json:"userSecret,omitempty" protobuf:"bytes,2,opt,name=user"`
	// Password for SASL/PLAIN authentication
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty" protobuf:"bytes,3,opt,name=password"`
}

// Backoff for an operation
type Backoff struct {
	// The initial duration in nanoseconds or strings like "1s", "3m"
	// +optional
	Duration *Int64OrString `json:"duration,omitempty" protobuf:"bytes,1,opt,name=duration"`
	// Duration is multiplied by factor each iteration
	// +optional
	Factor *Amount `json:"factor,omitempty" protobuf:"bytes,2,opt,name=factor"`
	// The amount of jitter applied each iteration
	// +optional
	Jitter *Amount `json:"jitter,omitempty" protobuf:"bytes,3,opt,name=jitter"`
	// Exit with error after this many steps
	// +optional
	Steps int32 `json:"steps,omitempty" protobuf:"varint,4,opt,name=steps"`
}

func (b Backoff) GetSteps() int {
	return int(b.Steps)
}

// Metadata holds the annotations and labels of an event source pod
type Metadata struct {
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,1,rep,name=annotations"`
	Labels      map[string]string `json:"labels,omitempty" protobuf:"bytes,2,rep,name=labels"`
}

func (s SASLConfig) GetMechanism() string {
	switch s.Mechanism {
	case "OAUTHBEARER", "SCRAM-SHA-256", "SCRAM-SHA-512", "GSSAPI":
		return s.Mechanism
	default:
		// default to PLAINTEXT mechanism
		return "PLAIN"
	}
}
