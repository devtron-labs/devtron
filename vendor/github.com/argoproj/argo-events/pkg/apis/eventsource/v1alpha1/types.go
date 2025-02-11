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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apicommon "github.com/argoproj/argo-events/pkg/apis/common"
)

// EventSource is the definition of a eventsource resource
// +genclient
// +kubebuilder:resource:shortName=es
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type EventSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Spec              EventSourceSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	// +optional
	Status EventSourceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// EventSourceList is the list of eventsource resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EventSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	Items []EventSource `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type EventSourceFilter struct {
	Expression string `json:"expression,omitempty" protobuf:"bytes,1,opt,name=expression"`
}

// EventSourceSpec refers to specification of event-source resource
type EventSourceSpec struct {
	// EventBusName references to a EventBus name. By default the value is "default"
	EventBusName string `json:"eventBusName,omitempty" protobuf:"bytes,1,opt,name=eventBusName"`
	// Template is the pod specification for the event source
	// +optional
	Template *Template `json:"template,omitempty" protobuf:"bytes,2,opt,name=template"`
	// Service is the specifications of the service to expose the event source
	// +optional
	Service *Service `json:"service,omitempty" protobuf:"bytes,3,opt,name=service"`
	// Minio event sources
	Minio map[string]apicommon.S3Artifact `json:"minio,omitempty" protobuf:"bytes,4,rep,name=minio"`
	// Calendar event sources
	Calendar map[string]CalendarEventSource `json:"calendar,omitempty" protobuf:"bytes,5,rep,name=calendar"`
	// File event sources
	File map[string]FileEventSource `json:"file,omitempty" protobuf:"bytes,6,rep,name=file"`
	// Resource event sources
	Resource map[string]ResourceEventSource `json:"resource,omitempty" protobuf:"bytes,7,rep,name=resource"`
	// Webhook event sources
	Webhook map[string]WebhookEventSource `json:"webhook,omitempty" protobuf:"bytes,8,rep,name=webhook"`
	// AMQP event sources
	AMQP map[string]AMQPEventSource `json:"amqp,omitempty" protobuf:"bytes,9,rep,name=amqp"`
	// Kafka event sources
	Kafka map[string]KafkaEventSource `json:"kafka,omitempty" protobuf:"bytes,10,rep,name=kafka"`
	// MQTT event sources
	MQTT map[string]MQTTEventSource `json:"mqtt,omitempty" protobuf:"bytes,11,rep,name=mqtt"`
	// NATS event sources
	NATS map[string]NATSEventsSource `json:"nats,omitempty" protobuf:"bytes,12,rep,name=nats"`
	// SNS event sources
	SNS map[string]SNSEventSource `json:"sns,omitempty" protobuf:"bytes,13,rep,name=sns"`
	// SQS event sources
	SQS map[string]SQSEventSource `json:"sqs,omitempty" protobuf:"bytes,14,rep,name=sqs"`
	// PubSub event sources
	PubSub map[string]PubSubEventSource `json:"pubSub,omitempty" protobuf:"bytes,15,rep,name=pubSub"`
	// Github event sources
	Github map[string]GithubEventSource `json:"github,omitempty" protobuf:"bytes,16,rep,name=github"`
	// Gitlab event sources
	Gitlab map[string]GitlabEventSource `json:"gitlab,omitempty" protobuf:"bytes,17,rep,name=gitlab"`
	// HDFS event sources
	HDFS map[string]HDFSEventSource `json:"hdfs,omitempty" protobuf:"bytes,18,rep,name=hdfs"`
	// Slack event sources
	Slack map[string]SlackEventSource `json:"slack,omitempty" protobuf:"bytes,19,rep,name=slack"`
	// StorageGrid event sources
	StorageGrid map[string]StorageGridEventSource `json:"storageGrid,omitempty" protobuf:"bytes,20,rep,name=storageGrid"`
	// AzureEventsHub event sources
	AzureEventsHub map[string]AzureEventsHubEventSource `json:"azureEventsHub,omitempty" protobuf:"bytes,21,rep,name=azureEventsHub"`
	// Stripe event sources
	Stripe map[string]StripeEventSource `json:"stripe,omitempty" protobuf:"bytes,22,rep,name=stripe"`
	// Emitter event source
	Emitter map[string]EmitterEventSource `json:"emitter,omitempty" protobuf:"bytes,23,rep,name=emitter"`
	// Redis event source
	Redis map[string]RedisEventSource `json:"redis,omitempty" protobuf:"bytes,24,rep,name=redis"`
	// NSQ event source
	NSQ map[string]NSQEventSource `json:"nsq,omitempty" protobuf:"bytes,25,rep,name=nsq"`
	// Pulsar event source
	Pulsar map[string]PulsarEventSource `json:"pulsar,omitempty" protobuf:"bytes,26,opt,name=pulsar"`
	// Generic event source
	Generic map[string]GenericEventSource `json:"generic,omitempty" protobuf:"bytes,27,rep,name=generic"`
	// Replicas is the event source deployment replicas
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,28,opt,name=replicas"`
	// Bitbucket Server event sources
	BitbucketServer map[string]BitbucketServerEventSource `json:"bitbucketserver,omitempty" protobuf:"bytes,29,rep,name=bitbucketserver"`
	// Bitbucket event sources
	Bitbucket map[string]BitbucketEventSource `json:"bitbucket,omitempty" protobuf:"bytes,30,rep,name=bitbucket"`
	// Redis stream source
	RedisStream map[string]RedisStreamEventSource `json:"redisStream,omitempty" protobuf:"bytes,31,rep,name=redisStream"`
}

func (e EventSourceSpec) GetReplicas() int32 {
	if e.Replicas == nil {
		return 1
	}
	var replicas int32
	if e.Replicas != nil {
		replicas = *e.Replicas
	}
	if replicas < 1 {
		replicas = 1
	}
	return replicas
}

// Template holds the information of an EventSource deployment template
type Template struct {
	// Metadata sets the pods's metadata, i.e. annotations and labels
	Metadata *apicommon.Metadata `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// ServiceAccountName is the name of the ServiceAccount to use to run event source pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,2,opt,name=serviceAccountName"`
	// Container is the main container image to run in the event source pod
	// +optional
	Container *corev1.Container `json:"container,omitempty" protobuf:"bytes,3,opt,name=container"`
	// Volumes is a list of volumes that can be mounted by containers in an eventsource.
	// +patchStrategy=merge
	// +patchMergeKey=name
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,4,rep,name=volumes"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty" protobuf:"bytes,5,opt,name=securityContext"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,6,opt,name=affinity"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,7,rep,name=tolerations"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,8,rep,name=nodeSelector"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,9,rep,name=imagePullSecrets"`
	// If specified, indicates the EventSource pod's priority. "system-node-critical"
	// and "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// More info: https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,10,opt,name=priorityClassName"`
	// The priority value. Various system components use this field to find the
	// priority of the EventSource pod. When Priority Admission Controller is enabled,
	// it prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// More info: https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/
	// +optional
	Priority *int32 `json:"priority,omitempty" protobuf:"bytes,11,opt,name=priority"`
}

// Service holds the service information eventsource exposes
type Service struct {
	// The list of ports that are exposed by this ClusterIP service.
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=protocol
	Ports []corev1.ServicePort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"port" protobuf:"bytes,1,rep,name=ports"`
	// clusterIP is the IP address of the service and is usually assigned
	// randomly by the master. If an address is specified manually and is not in
	// use by others, it will be allocated to the service; otherwise, creation
	// of the service will fail. This field can not be changed through updates.
	// Valid values are "None", empty string (""), or a valid IP address. "None"
	// can be specified for headless services when proxying is not required.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +optional
	ClusterIP string `json:"clusterIP,omitempty" protobuf:"bytes,2,opt,name=clusterIP"`
}

// CalendarEventSource describes an HTTP based EventSource
type WebhookEventSource struct {
	WebhookContext `json:",inline" protobuf:"bytes,1,opt,name=webhookContext"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,2,opt,name=filter"`
}

// CalendarEventSource describes a time based dependency. One of the fields (schedule, interval, or recurrence) must be passed.
// Schedule takes precedence over interval; interval takes precedence over recurrence
type CalendarEventSource struct {
	// Schedule is a cron-like expression. For reference, see: https://en.wikipedia.org/wiki/Cron
	// +optional
	Schedule string `json:"schedule" protobuf:"bytes,1,opt,name=schedule"`
	// Interval is a string that describes an interval duration, e.g. 1s, 30m, 2h...
	// +optional
	Interval string `json:"interval" protobuf:"bytes,2,opt,name=interval"`
	// ExclusionDates defines the list of DATE-TIME exceptions for recurring events.
	ExclusionDates []string `json:"exclusionDates,omitempty" protobuf:"bytes,3,rep,name=exclusionDates"`
	// Timezone in which to run the schedule
	// +optional
	Timezone string `json:"timezone,omitempty" protobuf:"bytes,4,opt,name=timezone"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,5,rep,name=metadata"`
	// Persistence hold the configuration for event persistence
	Persistence *EventPersistence `json:"persistence,omitempty" protobuf:"bytes,6,opt,name=persistence"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter"`
}

type EventPersistence struct {
	// Catchup enables to triggered the missed schedule when eventsource restarts
	Catchup *CatchupConfiguration `json:"catchup,omitempty" protobuf:"bytes,1,opt,name=catchup"`
	// ConfigMap holds configmap details for persistence
	ConfigMap *ConfigMapPersistence `json:"configMap,omitempty" protobuf:"bytes,2,opt,name=configMap"`
}

func (ep *EventPersistence) IsCatchUpEnabled() bool {
	return ep.Catchup != nil && ep.Catchup.Enabled
}

type CatchupConfiguration struct {
	// Enabled enables to triggered the missed schedule when eventsource restarts
	Enabled bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
	// MaxDuration holds max catchup duration
	MaxDuration string `json:"maxDuration,omitempty" protobuf:"bytes,2,opt,name=maxDuration"`
}

type ConfigMapPersistence struct {
	// Name of the configmap
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// CreateIfNotExist will create configmap if it doesn't exists
	CreateIfNotExist bool `json:"createIfNotExist,omitempty" protobuf:"varint,2,opt,name=createIfNotExist"`
}

// FileEventSource describes an event-source for file related events.
type FileEventSource struct {
	// Type of file operations to watch
	// Refer https://github.com/fsnotify/fsnotify/blob/master/fsnotify.go for more information
	EventType string `json:"eventType" protobuf:"bytes,1,opt,name=eventType"`
	// WatchPathConfig contains configuration about the file path to watch
	WatchPathConfig WatchPathConfig `json:"watchPathConfig" protobuf:"bytes,2,opt,name=watchPathConfig"`
	// Use polling instead of inotify
	Polling bool `json:"polling,omitempty" protobuf:"varint,3,opt,name=polling"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,4,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,5,opt,name=filter"`
}

// ResourceEventType is the type of event for the K8s resource mutation
type ResourceEventType string

// possible values of ResourceEventType
const (
	ADD    ResourceEventType = "ADD"
	UPDATE ResourceEventType = "UPDATE"
	DELETE ResourceEventType = "DELETE"
)

// ResourceEventSource refers to a event-source for K8s resource related events.
type ResourceEventSource struct {
	// Namespace where resource is deployed
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	// Filter is applied on the metadata of the resource
	// If you apply filter, then the internal event informer will only monitor objects that pass the filter.
	// +optional
	Filter *ResourceFilter `json:"filter,omitempty" protobuf:"bytes,2,opt,name=filter"`
	// Group of the resource
	metav1.GroupVersionResource `json:",inline" protobuf:"bytes,3,opt,name=groupVersionResource"`
	// EventTypes is the list of event type to watch.
	// Possible values are - ADD, UPDATE and DELETE.
	EventTypes []ResourceEventType `json:"eventTypes" protobuf:"bytes,4,rep,name=eventTypes,casttype=ResourceEventType"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,5,rep,name=metadata"`
}

// ResourceFilter contains K8s ObjectMeta information to further filter resource event objects
type ResourceFilter struct {
	// Prefix filter is applied on the resource name.
	// +optional
	Prefix string `json:"prefix,omitempty" protobuf:"bytes,1,opt,name=prefix"`
	// Labels provide listing options to K8s API to watch resource/s.
	// Refer https://kubernetes.io/docs/concepts/overview/working-with-objects/label-selectors/ for more info.
	// +optional
	Labels []Selector `json:"labels,omitempty" protobuf:"bytes,2,rep,name=labels"`
	// Fields provide field filters similar to K8s field selector
	// (see https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/).
	// Unlike K8s field selector, it supports arbitrary fileds like "spec.serviceAccountName",
	// and the value could be a string or a regex.
	// Same as K8s field selector, operator "=", "==" and "!=" are supported.
	// +optional
	Fields []Selector `json:"fields,omitempty" protobuf:"bytes,3,rep,name=fields"`
	// If resource is created before the specified time then the event is treated as valid.
	// +optional
	CreatedBy metav1.Time `json:"createdBy,omitempty" protobuf:"bytes,4,opt,name=createdBy"`
	// If the resource is created after the start time then the event is treated as valid.
	// +optional
	AfterStart bool `json:"afterStart,omitempty" protobuf:"varint,5,opt,name=afterStart"`
}

// Selector represents conditional operation to select K8s objects.
type Selector struct {
	// Key name
	Key string `json:"key" protobuf:"bytes,1,opt,name=key"`
	// Supported operations like ==, !=, <=, >= etc.
	// Defaults to ==.
	// Refer https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors for more info.
	// +optional
	Operation string `json:"operation,omitempty" protobuf:"bytes,2,opt,name=operation"`
	// Value
	Value string `json:"value" protobuf:"bytes,3,opt,name=value"`
}

// AMQPEventSource refers to an event-source for AMQP stream events
type AMQPEventSource struct {

	// URL for rabbitmq service
	URL string `json:"url,omitempty" protobuf:"bytes,1,opt,name=url"`
	// ExchangeName is the exchange name
	// For more information, visit https://www.rabbitmq.com/tutorials/amqp-concepts.html
	ExchangeName string `json:"exchangeName" protobuf:"bytes,2,opt,name=exchangeName"`
	// ExchangeType is rabbitmq exchange type
	ExchangeType string `json:"exchangeType" protobuf:"bytes,3,opt,name=exchangeType"`
	// Routing key for bindings
	RoutingKey string `json:"routingKey" protobuf:"bytes,4,opt,name=routingKey"`
	// Backoff holds parameters applied to connection.
	// +optional
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,5,opt,name=connectionBackoff"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,6,opt,name=jsonBody"`
	// TLS configuration for the amqp client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,7,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,8,rep,name=metadata"`
	// ExchangeDeclare holds the configuration for the exchange on the server
	// For more information, visit https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.ExchangeDeclare
	// +optional
	ExchangeDeclare *AMQPExchangeDeclareConfig `json:"exchangeDeclare,omitempty" protobuf:"bytes,9,opt,name=exchangeDeclare"`
	// QueueDeclare holds the configuration of a queue to hold messages and deliver to consumers.
	// Declaring creates a queue if it doesn't already exist, or ensures that an existing queue matches
	// the same parameters
	// For more information, visit https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.QueueDeclare
	// +optional
	QueueDeclare *AMQPQueueDeclareConfig `json:"queueDeclare,omitempty" protobuf:"bytes,10,opt,name=queueDeclare"`
	// QueueBind holds the configuration that binds an exchange to a queue so that publishings to the
	// exchange will be routed to the queue when the publishing routing key matches the binding routing key
	// For more information, visit https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.QueueBind
	// +optional
	QueueBind *AMQPQueueBindConfig `json:"queueBind,omitempty" protobuf:"bytes,11,opt,name=queueBind"`
	// Consume holds the configuration to immediately starts delivering queued messages
	// For more information, visit https://pkg.go.dev/github.com/rabbitmq/amqp091-go#Channel.Consume
	// +optional
	Consume *AMQPConsumeConfig `json:"consume,omitempty" protobuf:"bytes,12,opt,name=consume"`
	// Auth hosts secret selectors for username and password
	// +optional
	Auth *apicommon.BasicAuth `json:"auth,omitempty" protobuf:"bytes,13,opt,name=auth"`
	// URLSecret is secret reference for rabbitmq service URL
	URLSecret *corev1.SecretKeySelector `json:"urlSecret,omitempty" protobuf:"bytes,14,opt,name=urlSecret"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,15,opt,name=filter"`
}

// AMQPExchangeDeclareConfig holds the configuration for the exchange on the server
// +k8s:openapi-gen=true
type AMQPExchangeDeclareConfig struct {
	// Durable keeps the exchange also after the server restarts
	// +optional
	Durable bool `json:"durable,omitempty" protobuf:"varint,1,opt,name=durable"`
	// AutoDelete removes the exchange when no bindings are active
	// +optional
	AutoDelete bool `json:"autoDelete,omitempty" protobuf:"varint,2,opt,name=autoDelete"`
	// Internal when true does not accept publishings
	// +optional
	Internal bool `json:"internal,omitempty" protobuf:"varint,3,opt,name=internal"`
	// NowWait when true does not wait for a confirmation from the server
	// +optional
	NoWait bool `json:"noWait,omitempty" protobuf:"varint,4,opt,name=noWait"`
}

// AMQPQueueDeclareConfig holds the configuration of a queue to hold messages and deliver to consumers.
// Declaring creates a queue if it doesn't already exist, or ensures that an existing queue matches
// the same parameters
// +k8s:openapi-gen=true
type AMQPQueueDeclareConfig struct {
	// Name of the queue. If empty the server auto-generates a unique name for this queue
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// Durable keeps the queue also after the server restarts
	// +optional
	Durable bool `json:"durable,omitempty" protobuf:"varint,2,opt,name=durable"`
	// AutoDelete removes the queue when no consumers are active
	// +optional
	AutoDelete bool `json:"autoDelete,omitempty" protobuf:"varint,3,opt,name=autoDelete"`
	// Exclusive sets the queues to be accessible only by the connection that declares them and will be
	// deleted wgen the connection closes
	// +optional
	Exclusive bool `json:"exclusive,omitempty" protobuf:"varint,4,opt,name=exclusive"`
	// NowWait when true, the queue assumes to be declared on the server
	// +optional
	NoWait bool `json:"noWait,omitempty" protobuf:"varint,5,opt,name=noWait"`
	// Arguments of a queue (also known as "x-arguments") used for optional features and plugins
	// +optional
	Arguments string `json:"arguments,omitempty" protobuf:"bytes,6,opt,name=arguments"`
}

// AMQPQueueBindConfig holds the configuration that binds an exchange to a queue so that publishings to the
// exchange will be routed to the queue when the publishing routing key matches the binding routing key
// +k8s:openapi-gen=true
type AMQPQueueBindConfig struct {
	// NowWait false and the queue could not be bound, the channel will be closed with an error
	// +optional
	NoWait bool `json:"noWait,omitempty" protobuf:"varint,1,opt,name=noWait"`
}

// AMQPConsumeConfig holds the configuration to immediately starts delivering queued messages
// +k8s:openapi-gen=true
type AMQPConsumeConfig struct {
	// ConsumerTag is the identity of the consumer included in every delivery
	// +optional
	ConsumerTag string `json:"consumerTag,omitempty" protobuf:"bytes,1,opt,name=consumerTag"`
	// AutoAck when true, the server will acknowledge deliveries to this consumer prior to writing
	// the delivery to the network
	// +optional
	AutoAck bool `json:"autoAck,omitempty" protobuf:"varint,2,opt,name=autoAck"`
	// Exclusive when true, the server will ensure that this is the sole consumer from this queue
	// +optional
	Exclusive bool `json:"exclusive,omitempty" protobuf:"varint,3,opt,name=exclusive"`
	// NoLocal flag is not supported by RabbitMQ
	// +optional
	NoLocal bool `json:"noLocal,omitempty" protobuf:"varint,4,opt,name=noLocal"`
	// NowWait when true, do not wait for the server to confirm the request and immediately begin deliveries
	// +optional
	NoWait bool `json:"noWait,omitempty" protobuf:"varint,5,opt,name=noWait"`
}

// KafkaEventSource refers to event-source for Kafka related events
type KafkaEventSource struct {
	// URL to kafka cluster, multiple URLs separated by comma
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
	// Partition name
	Partition string `json:"partition" protobuf:"bytes,2,opt,name=partition"`
	// Topic name
	Topic string `json:"topic" protobuf:"bytes,3,opt,name=topic"`
	// Backoff holds parameters applied to connection.
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,4,opt,name=connectionBackoff"`
	// TLS configuration for the kafka client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,5,opt,name=tls"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,6,opt,name=jsonBody"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,7,rep,name=metadata"`

	// Consumer group for kafka client
	// +optional
	ConsumerGroup *KafkaConsumerGroup `json:"consumerGroup,omitempty" protobuf:"bytes,8,opt,name=consumerGroup"`

	// Sets a limit on how many events get read from kafka per second.
	// +optional
	LimitEventsPerSecond int64 `json:"limitEventsPerSecond,omitempty" protobuf:"varint,9,opt,name=limitEventsPerSecond"`

	// Specify what kafka version is being connected to enables certain features in sarama, defaults to 1.0.0
	// +optional
	Version string `json:"version" protobuf:"bytes,10,opt,name=version"`
	// SASL configuration for the kafka client
	// +optional
	SASL *apicommon.SASLConfig `json:"sasl,omitempty" protobuf:"bytes,11,opt,name=sasl"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,12,opt,name=filter"`
	// Yaml format Sarama config for Kafka connection.
	// It follows the struct of sarama.Config. See https://github.com/Shopify/sarama/blob/main/config.go
	// e.g.
	//
	// consumer:
	//   fetch:
	//     min: 1
	// net:
	//   MaxOpenRequests: 5
	//
	// +optional
	Config string `json:"config,omitempty" protobuf:"bytes,13,opt,name=config"`
}

type KafkaConsumerGroup struct {
	// The name for the consumer group to use
	GroupName string `json:"groupName" protobuf:"bytes,1,opt,name=groupName"`
	// When starting up a new group do we want to start from the oldest event (true) or the newest event (false), defaults to false
	// +optional
	Oldest bool `json:"oldest,omitempty" protobuf:"varint,2,opt,name=oldest"`
	// Rebalance strategy can be one of: sticky, roundrobin, range. Range is the default.
	// +optional
	RebalanceStrategy string `json:"rebalanceStrategy" protobuf:"bytes,3,opt,name=rebalanceStrategy"`
}

// MQTTEventSource refers to event-source for MQTT related events
type MQTTEventSource struct {
	// URL to connect to broker
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
	// Topic name
	Topic string `json:"topic" protobuf:"bytes,2,opt,name=topic"`
	// ClientID is the id of the client
	ClientID string `json:"clientId" protobuf:"bytes,3,opt,name=clientId"`
	// ConnectionBackoff holds backoff applied to connection.
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,4,opt,name=connectionBackoff"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,5,opt,name=jsonBody"`
	// TLS configuration for the mqtt client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,7,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter"`
}

// NATSEventsSource refers to event-source for NATS related events
type NATSEventsSource struct {
	// URL to connect to NATS cluster
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
	// Subject holds the name of the subject onto which messages are published
	Subject string `json:"subject" protobuf:"bytes,2,opt,name=subject"`
	// ConnectionBackoff holds backoff applied to connection.
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,3,opt,name=connectionBackoff"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,4,opt,name=jsonBody"`
	// TLS configuration for the nats client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,5,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,6,rep,name=metadata"`
	// Auth information
	// +optional
	Auth *NATSAuth `json:"auth,omitempty" protobuf:"bytes,7,opt,name=auth"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter"`
}

// NATSAuth refers to the auth info for NATS EventSource
type NATSAuth struct {
	// Baisc auth with username and password
	// +optional
	Basic *apicommon.BasicAuth `json:"basic,omitempty" protobuf:"bytes,1,opt,name=basic"`
	// Token used to connect
	// +optional
	Token *corev1.SecretKeySelector `json:"token,omitempty" protobuf:"bytes,2,opt,name=token"`
	// NKey used to connect
	// +optional
	NKey *corev1.SecretKeySelector `json:"nkey,omitempty" protobuf:"bytes,3,opt,name=nkey"`
	// credential used to connect
	// +optional
	Credential *corev1.SecretKeySelector `json:"credential,omitempty" protobuf:"bytes,4,opt,name=credential"`
}

// SNSEventSource refers to event-source for AWS SNS related events
type SNSEventSource struct {
	// Webhook configuration for http server
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,1,opt,name=webhook"`
	// TopicArn
	TopicArn string `json:"topicArn" protobuf:"bytes,2,opt,name=topicArn"`
	// AccessKey refers K8s secret containing aws access key
	AccessKey *corev1.SecretKeySelector `json:"accessKey,omitempty" protobuf:"bytes,3,opt,name=accessKey"`
	// SecretKey refers K8s secret containing aws secret key
	SecretKey *corev1.SecretKeySelector `json:"secretKey,omitempty" protobuf:"bytes,4,opt,name=secretKey"`
	// Region is AWS region
	Region string `json:"region" protobuf:"bytes,5,opt,name=region"`
	// RoleARN is the Amazon Resource Name (ARN) of the role to assume.
	// +optional
	RoleARN string `json:"roleARN,omitempty" protobuf:"bytes,6,opt,name=roleARN"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,7,rep,name=metadata"`
	// ValidateSignature is boolean that can be set to true for SNS signature verification
	// +optional
	ValidateSignature bool `json:"validateSignature,omitempty" protobuf:"varint,8,opt,name=validateSignature"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,9,opt,name=filter"`
	// Endpoint configures connection to a specific SNS endpoint instead of Amazons servers
	// +optional
	Endpoint string `json:"endpoint" protobuf:"bytes,10,opt,name=endpoint"`
}

// SQSEventSource refers to event-source for AWS SQS related events
type SQSEventSource struct {
	// AccessKey refers K8s secret containing aws access key
	AccessKey *corev1.SecretKeySelector `json:"accessKey,omitempty" protobuf:"bytes,1,opt,name=accessKey"`
	// SecretKey refers K8s secret containing aws secret key
	SecretKey *corev1.SecretKeySelector `json:"secretKey,omitempty" protobuf:"bytes,2,opt,name=secretKey"`
	// Region is AWS region
	Region string `json:"region" protobuf:"bytes,3,opt,name=region"`
	// Queue is AWS SQS queue to listen to for messages
	Queue string `json:"queue" protobuf:"bytes,4,opt,name=queue"`
	// WaitTimeSeconds is The duration (in seconds) for which the call waits for a message to arrive
	// in the queue before returning.
	WaitTimeSeconds int64 `json:"waitTimeSeconds" protobuf:"varint,5,opt,name=waitTimeSeconds"`
	// RoleARN is the Amazon Resource Name (ARN) of the role to assume.
	// +optional
	RoleARN string `json:"roleARN,omitempty" protobuf:"bytes,6,opt,name=roleARN"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,7,opt,name=jsonBody"`
	// QueueAccountID is the ID of the account that created the queue to monitor
	// +optional
	QueueAccountID string `json:"queueAccountId,omitempty" protobuf:"bytes,8,opt,name=queueAccountId"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,9,rep,name=metadata"`
	// DLQ specifies if a dead-letter queue is configured for messages that can't be processed successfully.
	// If set to true, messages with invalid payload won't be acknowledged to allow to forward them farther to the dead-letter queue.
	// The default value is false.
	// +optional
	DLQ bool `json:"dlq,omitempty" protobuf:"varint,10,opt,name=dlq"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,11,opt,name=filter"`
	// Endpoint configures connection to a specific SQS endpoint instead of Amazons servers
	// +optional
	Endpoint string `json:"endpoint" protobuf:"bytes,12,opt,name=endpoint"`
	// SessionToken refers to K8s secret containing AWS temporary credentials(STS) session token
	// +optional
	SessionToken *corev1.SecretKeySelector `json:"sessionToken,omitempty" protobuf:"bytes,13,opt,name=sessionToken"`
}

// PubSubEventSource refers to event-source for GCP PubSub related events.
type PubSubEventSource struct {
	// ProjectID is GCP project ID for the subscription.
	// Required if you run Argo Events outside of GKE/GCE.
	// (otherwise, the default value is its project)
	// +optional
	ProjectID string `json:"projectID" protobuf:"bytes,1,opt,name=projectID"`
	// TopicProjectID is GCP project ID for the topic.
	// By default, it is same as ProjectID.
	// +optional
	TopicProjectID string `json:"topicProjectID" protobuf:"bytes,2,opt,name=topicProjectID"`
	// Topic to which the subscription should belongs.
	// Required if you want the eventsource to create a new subscription.
	// If you specify this field along with an existing subscription,
	// it will be verified whether it actually belongs to the specified topic.
	// +optional
	Topic string `json:"topic" protobuf:"bytes,3,opt,name=topic"`
	// SubscriptionID is ID of subscription.
	// Required if you use existing subscription.
	// The default value will be auto generated hash based on this eventsource setting, so the subscription
	// might be recreated every time you update the setting, which has a possibility of event loss.
	// +optional
	SubscriptionID string `json:"subscriptionID" protobuf:"bytes,4,opt,name=subscriptionID"`
	// CredentialSecret references to the secret that contains JSON credentials to access GCP.
	// If it is missing, it implicitly uses Workload Identity to access.
	// https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
	// +optional
	CredentialSecret *corev1.SecretKeySelector `json:"credentialSecret,omitempty" protobuf:"bytes,5,opt,name=credentialSecret"`
	// DeleteSubscriptionOnFinish determines whether to delete the GCP PubSub subscription once the event source is stopped.
	// +optional
	DeleteSubscriptionOnFinish bool `json:"deleteSubscriptionOnFinish,omitempty" protobuf:"varint,6,opt,name=deleteSubscriptionOnFinish"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,7,opt,name=jsonBody"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,8,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,9,opt,name=filter"`
}

type OwnedRepositories struct {
	// Organization or user name
	Owner string `json:"owner,omitempty" protobuf:"bytes,1,opt,name=owner"`
	// Repository names
	Names []string `json:"names,omitempty" protobuf:"bytes,2,rep,name=names"`
}

type GithubAppCreds struct {
	// PrivateKey refers to a K8s secret containing the GitHub app private key
	PrivateKey *corev1.SecretKeySelector `json:"privateKey" protobuf:"bytes,1,opt,name=privateKey"`
	// AppID refers to the GitHub App ID for the application you created
	AppID int64 `json:"appID" protobuf:"bytes,2,opt,name=appID"`
	// InstallationID refers to the Installation ID of the GitHub app you created and installed
	InstallationID int64 `json:"installationID" protobuf:"bytes,3,opt,name=installationID"`
}

// GithubEventSource refers to event-source for github related events
type GithubEventSource struct {
	// Id is the webhook's id
	// Deprecated: This is not used at all, will be removed in v1.6
	// +optional
	ID int64 `json:"id" protobuf:"varint,1,opt,name=id"`
	// Webhook refers to the configuration required to run a http server
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,2,opt,name=webhook"`
	// DeprecatedOwner refers to GitHub owner name i.e. argoproj
	// Deprecated: use Repositories instead. Will be unsupported in v 1.6
	// +optional
	DeprecatedOwner string `json:"owner" protobuf:"bytes,3,opt,name=owner"`
	// DeprecatedRepository refers to GitHub repo name i.e. argo-events
	// Deprecated: use Repositories instead. Will be unsupported in v 1.6
	// +optional
	DeprecatedRepository string `json:"repository" protobuf:"bytes,4,opt,name=repository"`
	// Events refer to Github events to which the event source will subscribe
	Events []string `json:"events" protobuf:"bytes,5,rep,name=events"`
	// APIToken refers to a K8s secret containing github api token
	// +optional
	APIToken *corev1.SecretKeySelector `json:"apiToken,omitempty" protobuf:"bytes,6,opt,name=apiToken"`
	// WebhookSecret refers to K8s secret containing GitHub webhook secret
	// https://developer.github.com/webhooks/securing/
	// +optional
	WebhookSecret *corev1.SecretKeySelector `json:"webhookSecret,omitempty" protobuf:"bytes,7,opt,name=webhookSecret"`
	// Insecure tls verification
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,8,opt,name=insecure"`
	// Active refers to status of the webhook for event deliveries.
	// https://developer.github.com/webhooks/creating/#active
	// +optional
	Active bool `json:"active,omitempty" protobuf:"varint,9,opt,name=active"`
	// ContentType of the event delivery
	ContentType string `json:"contentType,omitempty" protobuf:"bytes,10,opt,name=contentType"`
	// GitHub base URL (for GitHub Enterprise)
	// +optional
	GithubBaseURL string `json:"githubBaseURL,omitempty" protobuf:"bytes,11,opt,name=githubBaseURL"`
	// GitHub upload URL (for GitHub Enterprise)
	// +optional
	GithubUploadURL string `json:"githubUploadURL,omitempty" protobuf:"bytes,12,opt,name=githubUploadURL"`
	// DeleteHookOnFinish determines whether to delete the GitHub hook for the repository once the event source is stopped.
	// +optional
	DeleteHookOnFinish bool `json:"deleteHookOnFinish,omitempty" protobuf:"varint,13,opt,name=deleteHookOnFinish"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,14,rep,name=metadata"`
	// Repositories holds the information of repositories, which uses repo owner as the key,
	// and list of repo names as the value. Not required if Organizations is set.
	Repositories []OwnedRepositories `json:"repositories,omitempty" protobuf:"bytes,15,rep,name=repositories"`
	// Organizations holds the names of organizations (used for organization level webhooks). Not required if Repositories is set.
	Organizations []string `json:"organizations,omitempty" protobuf:"bytes,16,rep,name=organizations"`
	// GitHubApp holds the GitHub app credentials
	// +optional
	GithubApp *GithubAppCreds `json:"githubApp,omitempty" protobuf:"bytes,17,opt,name=githubApp"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,18,opt,name=filter"`
}

func (g GithubEventSource) GetOwnedRepositories() []OwnedRepositories {
	if len(g.Repositories) > 0 {
		return g.Repositories
	} else if g.DeprecatedOwner != "" && g.DeprecatedRepository != "" {
		return []OwnedRepositories{
			{
				Owner: g.DeprecatedOwner,
				Names: []string{
					g.DeprecatedRepository,
				},
			},
		}
	}
	return nil
}

func (g GithubEventSource) HasGithubAPIToken() bool {
	return g.APIToken != nil
}

func (g GithubEventSource) HasGithubAppCreds() bool {
	return g.GithubApp != nil && g.GithubApp.PrivateKey != nil
}

func (g GithubEventSource) HasConfiguredWebhook() bool {
	return g.Webhook != nil && g.Webhook.URL != ""
}

func (g GithubEventSource) NeedToCreateHooks() bool {
	return (g.HasGithubAPIToken() || g.HasGithubAppCreds()) && g.HasConfiguredWebhook()
}

// GitlabEventSource refers to event-source related to Gitlab events
type GitlabEventSource struct {
	// Webhook holds configuration to run a http server
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,1,opt,name=webhook"`
	// DeprecatedProjectID is the id of project for which integration needs to setup
	// Deprecated: use Projects instead. Will be unsupported in v 1.7
	// +optional
	DeprecatedProjectID string `json:"projectID,omitempty" protobuf:"bytes,2,opt,name=projectID"`
	// Events are gitlab event to listen to.
	// Refer https://github.com/xanzy/go-gitlab/blob/bf34eca5d13a9f4c3f501d8a97b8ac226d55e4d9/projects.go#L794.
	Events []string `json:"events" protobuf:"bytes,3,opt,name=events"`
	// AccessToken references to k8 secret which holds the gitlab api access information
	AccessToken *corev1.SecretKeySelector `json:"accessToken,omitempty" protobuf:"bytes,4,opt,name=accessToken"`
	// EnableSSLVerification to enable ssl verification
	// +optional
	EnableSSLVerification bool `json:"enableSSLVerification,omitempty" protobuf:"varint,5,opt,name=enableSSLVerification"`
	// GitlabBaseURL is the base URL for API requests to a custom endpoint
	GitlabBaseURL string `json:"gitlabBaseURL" protobuf:"bytes,6,opt,name=gitlabBaseURL"`
	// DeleteHookOnFinish determines whether to delete the GitLab hook for the project once the event source is stopped.
	// +optional
	DeleteHookOnFinish bool `json:"deleteHookOnFinish,omitempty" protobuf:"varint,8,opt,name=deleteHookOnFinish"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,9,rep,name=metadata"`
	// List of project IDs or project namespace paths like "whynowy/test"
	Projects []string `json:"projects,omitempty" protobuf:"bytes,10,rep,name=projects"`
	// SecretToken references to k8 secret which holds the Secret Token used by webhook config
	SecretToken *corev1.SecretKeySelector `json:"secretToken,omitempty" protobuf:"bytes,11,opt,name=secretToken"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,12,opt,name=filter"`
}

func (g GitlabEventSource) GetProjects() []string {
	if len(g.Projects) > 0 {
		return g.Projects
	}
	if g.DeprecatedProjectID != "" {
		return []string{g.DeprecatedProjectID}
	}
	return []string{}
}

func (g GitlabEventSource) NeedToCreateHooks() bool {
	return g.AccessToken != nil && g.Webhook != nil && g.Webhook.URL != ""
}

// BitbucketEventSource describes the event source for Bitbucket
type BitbucketEventSource struct {
	// DeleteHookOnFinish determines whether to delete the defined Bitbucket hook once the event source is stopped.
	// +optional
	DeleteHookOnFinish bool `json:"deleteHookOnFinish,omitempty" protobuf:"varint,1,opt,name=deleteHookOnFinish"`
	// Metadata holds the user defined metadata which will be passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,2,rep,name=metadata"`
	// Webhook refers to the configuration required to run an http server
	Webhook *WebhookContext `json:"webhook" protobuf:"bytes,3,name=webhook"`
	// Auth information required to connect to Bitbucket.
	Auth *BitbucketAuth `json:"auth" protobuf:"bytes,4,name=auth"`
	// Events this webhook is subscribed to.
	Events []string `json:"events" protobuf:"bytes,5,name=events"`
	// DeprecatedOwner is the owner of the repository.
	// Deprecated: use Repositories instead. Will be unsupported in v1.9
	// +optional
	DeprecatedOwner string `json:"owner,omitempty" protobuf:"bytes,6,name=owner"`
	// DeprecatedProjectKey is the key of the project to which the repository relates
	// Deprecated: use Repositories instead. Will be unsupported in v1.9
	// +optional
	DeprecatedProjectKey string `json:"projectKey,omitempty" protobuf:"bytes,7,opt,name=projectKey"`
	// DeprecatedRepositorySlug is a URL-friendly version of a repository name, automatically generated by Bitbucket for use in the URL
	// Deprecated: use Repositories instead. Will be unsupported in v1.9
	// +optional
	DeprecatedRepositorySlug string `json:"repositorySlug,omitempty" protobuf:"bytes,8,name=repositorySlug"`
	// Repositories holds a list of repositories for which integration needs to set up
	// +optional
	Repositories []BitbucketRepository `json:"repositories,omitempty" protobuf:"bytes,9,rep,name=repositories"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,10,opt,name=filter"`
}

func (b BitbucketEventSource) HasBitbucketBasicAuth() bool {
	return b.Auth.Basic != nil && b.Auth.Basic.Username != nil && b.Auth.Basic.Password != nil
}

func (b BitbucketEventSource) HasBitbucketOAuthToken() bool {
	return b.Auth.OAuthToken != nil
}

func (b BitbucketEventSource) HasConfiguredWebhook() bool {
	return b.Webhook != nil && b.Webhook.URL != ""
}

func (b BitbucketEventSource) ShouldCreateWebhooks() bool {
	return (b.HasBitbucketBasicAuth() || b.HasBitbucketOAuthToken()) && b.HasConfiguredWebhook()
}

func (b BitbucketEventSource) GetBitbucketRepositories() []BitbucketRepository {
	if len(b.Repositories) > 0 {
		return b.Repositories
	}

	if b.DeprecatedOwner != "" && b.DeprecatedRepositorySlug != "" {
		return []BitbucketRepository{
			{
				Owner:          b.DeprecatedOwner,
				RepositorySlug: b.DeprecatedRepositorySlug,
			},
		}
	}

	return nil
}

type BitbucketRepository struct {
	// Owner is the owner of the repository
	Owner string `json:"owner" protobuf:"bytes,1,name=owner"`
	// RepositorySlug is a URL-friendly version of a repository name, automatically generated by Bitbucket for use in the URL
	RepositorySlug string `json:"repositorySlug" protobuf:"bytes,2,rep,name=repositorySlug"`
}

// GetRepositoryID helper returns a string key identifier for the repo
func (r BitbucketRepository) GetRepositoryID() string {
	return r.Owner + "," + r.RepositorySlug
}

// BitbucketAuth holds the different auth strategies for connecting to Bitbucket
type BitbucketAuth struct {
	// Basic is BasicAuth auth strategy.
	// +optional
	Basic *BitbucketBasicAuth `json:"basic,omitempty" protobuf:"bytes,1,opt,name=basic"`
	// OAuthToken refers to the K8s secret that holds the OAuth Bearer token.
	// +optional
	OAuthToken *corev1.SecretKeySelector `json:"oauthToken,omitempty" protobuf:"bytes,2,opt,name=oauthToken"`
}

// BasicAuth holds the information required to authenticate user via basic auth mechanism
type BitbucketBasicAuth struct {
	// Username refers to the K8s secret that holds the username.
	Username *corev1.SecretKeySelector `json:"username" protobuf:"bytes,1,name=username"`
	// Password refers to the K8s secret that holds the password.
	Password *corev1.SecretKeySelector `json:"password" protobuf:"bytes,2,name=password"`
}

// BitbucketServerEventSource refers to event-source related to Bitbucket Server events
type BitbucketServerEventSource struct {
	// Webhook holds configuration to run a http server
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,1,opt,name=webhook"`
	// DeprecatedProjectKey is the key of project for which integration needs to set up
	// Deprecated: use Repositories instead. Will be unsupported in v1.8
	// +optional
	DeprecatedProjectKey string `json:"projectKey,omitempty" protobuf:"bytes,2,opt,name=projectKey"`
	// DeprecatedRepositorySlug is the slug of the repository for which integration needs to set up
	// Deprecated: use Repositories instead. Will be unsupported in v1.8
	// +optional
	DeprecatedRepositorySlug string `json:"repositorySlug,omitempty" protobuf:"bytes,3,opt,name=repositorySlug"`
	// Repositories holds a list of repositories for which integration needs to set up
	// +optional
	Repositories []BitbucketServerRepository `json:"repositories,omitempty" protobuf:"bytes,4,rep,name=repositories"`
	// Events are bitbucket event to listen to.
	// Refer https://confluence.atlassian.com/bitbucketserver/event-payload-938025882.html
	Events []string `json:"events" protobuf:"bytes,5,opt,name=events"`
	// AccessToken is reference to K8s secret which holds the bitbucket api access information
	AccessToken *corev1.SecretKeySelector `json:"accessToken,omitempty" protobuf:"bytes,6,opt,name=accessToken"`
	// WebhookSecret is reference to K8s secret which holds the bitbucket webhook secret (for HMAC validation)
	WebhookSecret *corev1.SecretKeySelector `json:"webhookSecret,omitempty" protobuf:"bytes,7,opt,name=webhookSecret"`
	// BitbucketServerBaseURL is the base URL for API requests to a custom endpoint
	BitbucketServerBaseURL string `json:"bitbucketserverBaseURL" protobuf:"bytes,8,opt,name=bitbucketserverBaseURL"`
	// DeleteHookOnFinish determines whether to delete the Bitbucket Server hook for the project once the event source is stopped.
	// +optional
	DeleteHookOnFinish bool `json:"deleteHookOnFinish,omitempty" protobuf:"varint,9,opt,name=deleteHookOnFinish"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,10,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,11,opt,name=filter"`
}

type BitbucketServerRepository struct {
	// ProjectKey is the key of project for which integration needs to set up
	ProjectKey string `json:"projectKey" protobuf:"bytes,1,opt,name=projectKey"`
	// RepositorySlug is the slug of the repository for which integration needs to set up
	RepositorySlug string `json:"repositorySlug" protobuf:"bytes,2,rep,name=repositorySlug"`
}

func (b BitbucketServerEventSource) ShouldCreateWebhooks() bool {
	return b.AccessToken != nil && b.Webhook != nil && b.Webhook.URL != ""
}

func (b BitbucketServerEventSource) GetBitbucketServerRepositories() []BitbucketServerRepository {
	if len(b.Repositories) > 0 {
		return b.Repositories
	}

	if b.DeprecatedProjectKey != "" && b.DeprecatedRepositorySlug != "" {
		return []BitbucketServerRepository{
			{
				ProjectKey:     b.DeprecatedProjectKey,
				RepositorySlug: b.DeprecatedRepositorySlug,
			},
		}
	}

	return nil
}

// HDFSEventSource refers to event-source for HDFS related events
type HDFSEventSource struct {
	WatchPathConfig `json:",inline" protobuf:"bytes,1,opt,name=watchPathConfig"`
	// Type of file operations to watch
	Type string `json:"type" protobuf:"bytes,2,opt,name=type"`
	// CheckInterval is a string that describes an interval duration to check the directory state, e.g. 1s, 30m, 2h... (defaults to 1m)
	CheckInterval string `json:"checkInterval,omitempty" protobuf:"bytes,3,opt,name=checkInterval"`
	// Addresses is accessible addresses of HDFS name nodes

	Addresses []string `json:"addresses" protobuf:"bytes,4,rep,name=addresses"`
	// HDFSUser is the user to access HDFS file system.
	// It is ignored if either ccache or keytab is used.
	HDFSUser string `json:"hdfsUser,omitempty" protobuf:"bytes,5,opt,name=hdfsUser"`
	// KrbCCacheSecret is the secret selector for Kerberos ccache
	// Either ccache or keytab can be set to use Kerberos.
	KrbCCacheSecret *corev1.SecretKeySelector `json:"krbCCacheSecret,omitempty" protobuf:"bytes,6,opt,name=krbCCacheSecret"`
	// KrbKeytabSecret is the secret selector for Kerberos keytab
	// Either ccache or keytab can be set to use Kerberos.
	KrbKeytabSecret *corev1.SecretKeySelector `json:"krbKeytabSecret,omitempty" protobuf:"bytes,7,opt,name=krbKeytabSecret"`
	// KrbUsername is the Kerberos username used with Kerberos keytab
	// It must be set if keytab is used.
	KrbUsername string `json:"krbUsername,omitempty" protobuf:"bytes,8,opt,name=krbUsername"`
	// KrbRealm is the Kerberos realm used with Kerberos keytab
	// It must be set if keytab is used.
	KrbRealm string `json:"krbRealm,omitempty" protobuf:"bytes,9,opt,name=krbRealm"`
	// KrbConfig is the configmap selector for Kerberos config as string
	// It must be set if either ccache or keytab is used.
	KrbConfigConfigMap *corev1.ConfigMapKeySelector `json:"krbConfigConfigMap,omitempty" protobuf:"bytes,10,opt,name=krbConfigConfigMap"`
	// KrbServicePrincipalName is the principal name of Kerberos service
	// It must be set if either ccache or keytab is used.
	KrbServicePrincipalName string `json:"krbServicePrincipalName,omitempty" protobuf:"bytes,11,opt,name=krbServicePrincipalName"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,12,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,13,opt,name=filter"`
}

// SlackEventSource refers to event-source for Slack related events
type SlackEventSource struct {
	// Slack App signing secret
	SigningSecret *corev1.SecretKeySelector `json:"signingSecret,omitempty" protobuf:"bytes,1,opt,name=signingSecret"`
	// Token for URL verification handshake
	Token *corev1.SecretKeySelector `json:"token,omitempty" protobuf:"bytes,2,opt,name=token"`
	// Webhook holds configuration for a REST endpoint
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,3,opt,name=webhook"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,4,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,5,opt,name=filter"`
}

// StorageGridEventSource refers to event-source for StorageGrid related events
type StorageGridEventSource struct {
	// Webhook holds configuration for a REST endpoint
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,1,opt,name=webhook"`
	// Events are s3 bucket notification events.
	// For more information on s3 notifications, follow https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
	// Note that storage grid notifications do not contain `s3:`

	Events []string `json:"events,omitempty" protobuf:"bytes,2,rep,name=events"`
	// Filter on object key which caused the notification.
	Filter *StorageGridFilter `json:"filter,omitempty" protobuf:"bytes,3,opt,name=filter"`
	// TopicArn
	TopicArn string `json:"topicArn" protobuf:"bytes,4,name=topicArn"`
	// Name of the bucket to register notifications for.
	Bucket string `json:"bucket" protobuf:"bytes,5,name=bucket"`
	// S3 region.
	// Defaults to us-east-1
	// +optional
	Region string `json:"region,omitempty" protobuf:"bytes,6,opt,name=region"`
	// Auth token for storagegrid api
	AuthToken *corev1.SecretKeySelector `json:"authToken" protobuf:"bytes,7,name=authToken"`
	// APIURL is the url of the storagegrid api.
	APIURL string `json:"apiURL" protobuf:"bytes,8,name=apiURL"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,9,rep,name=metadata"`
}

// StorageGridFilter represents filters to apply to bucket notifications for specifying constraints on objects
// +k8s:openapi-gen=true
type StorageGridFilter struct {
	Prefix string `json:"prefix" protobuf:"bytes,1,opt,name=prefix"`
	Suffix string `json:"suffix" protobuf:"bytes,2,opt,name=suffix"`
}

// AzureEventsHubEventSource describes the event source for azure events hub
// More info at https://docs.microsoft.com/en-us/azure/event-hubs/
type AzureEventsHubEventSource struct {
	// FQDN of the EventHubs namespace you created
	// More info at https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-get-connection-string
	FQDN string `json:"fqdn" protobuf:"bytes,1,opt,name=fqdn"`
	// SharedAccessKeyName is the name you chose for your application's SAS keys
	SharedAccessKeyName *corev1.SecretKeySelector `json:"sharedAccessKeyName,omitempty" protobuf:"bytes,2,opt,name=sharedAccessKeyName"`
	// SharedAccessKey is the generated value of the key
	SharedAccessKey *corev1.SecretKeySelector `json:"sharedAccessKey,omitempty" protobuf:"bytes,3,opt,name=sharedAccessKey"`
	// Event Hub path/name
	HubName string `json:"hubName" protobuf:"bytes,4,opt,name=hubName"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,5,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,6,opt,name=filter"`
}

// StripeEventSource describes the event source for stripe webhook notifications
// More info at https://stripe.com/docs/webhooks
type StripeEventSource struct {
	// Webhook holds configuration for a REST endpoint
	Webhook *WebhookContext `json:"webhook,omitempty" protobuf:"bytes,1,opt,name=webhook"`
	// CreateWebhook if specified creates a new webhook programmatically.
	// +optional
	CreateWebhook bool `json:"createWebhook,omitempty" protobuf:"varint,2,opt,name=createWebhook"`
	// APIKey refers to K8s secret that holds Stripe API key. Used only if CreateWebhook is enabled.
	// +optional
	APIKey *corev1.SecretKeySelector `json:"apiKey,omitempty" protobuf:"bytes,3,opt,name=apiKey"`
	// EventFilter describes the type of events to listen to. If not specified, all types of events will be processed.
	// More info at https://stripe.com/docs/api/events/list
	// +optional
	EventFilter []string `json:"eventFilter,omitempty" protobuf:"bytes,4,rep,name=eventFilter"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,5,rep,name=metadata"`
}

// EmitterEventSource describes the event source for emitter
// More info at https://emitter.io/develop/getting-started/
type EmitterEventSource struct {
	// Broker URI to connect to.
	Broker string `json:"broker" protobuf:"bytes,1,opt,name=broker"`
	// ChannelKey refers to the channel key
	ChannelKey string `json:"channelKey" protobuf:"bytes,2,opt,name=channelKey"`
	// ChannelName refers to the channel name
	ChannelName string `json:"channelName" protobuf:"bytes,3,opt,name=channelName"`
	// Username to use to connect to broker
	// +optional
	Username *corev1.SecretKeySelector `json:"username,omitempty" protobuf:"bytes,4,opt,name=username"`
	// Password to use to connect to broker
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty" protobuf:"bytes,5,opt,name=password"`
	// Backoff holds parameters applied to connection.
	// +optional
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,6,opt,name=connectionBackoff"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,7,opt,name=jsonBody"`
	// TLS configuration for the emitter client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,8,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,9,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,10,opt,name=filter"`
}

// RedisEventSource describes an event source for the Redis PubSub.
// More info at https://godoc.org/github.com/go-redis/redis#example-PubSub
type RedisEventSource struct {
	// HostAddress refers to the address of the Redis host/server
	HostAddress string `json:"hostAddress" protobuf:"bytes,1,opt,name=hostAddress"`
	// Password required for authentication if any.
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty" protobuf:"bytes,2,opt,name=password"`
	// Namespace to use to retrieve the password from. It should only be specified if password is declared
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	// DB to use. If not specified, default DB 0 will be used.
	// +optional
	DB int32 `json:"db,omitempty" protobuf:"varint,4,opt,name=db"`
	// Channels to subscribe to listen events.

	Channels []string `json:"channels" protobuf:"bytes,5,rep,name=channels"`
	// TLS configuration for the redis client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,7,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,9,opt,name=jsonBody"`
	// Username required for ACL style authentication if any.
	// +optional
	Username string `json:"username,omitempty" protobuf:"bytes,10,opt,name=username"`
}

// RedisStreamEventSource describes an event source for
// Redis streams (https://redis.io/topics/streams-intro)
type RedisStreamEventSource struct {
	// HostAddress refers to the address of the Redis host/server (master instance)
	HostAddress string `json:"hostAddress" protobuf:"bytes,1,opt,name=hostAddress"`
	// Password required for authentication if any.
	// +optional
	Password *corev1.SecretKeySelector `json:"password,omitempty" protobuf:"bytes,2,opt,name=password"`
	// DB to use. If not specified, default DB 0 will be used.
	// +optional
	DB int32 `json:"db,omitempty" protobuf:"varint,3,opt,name=db"`
	// Streams to look for entries. XREADGROUP is used on all streams using a single consumer group.
	Streams []string `json:"streams" protobuf:"bytes,4,rep,name=streams"`
	// MaxMsgCountPerRead holds the maximum number of messages per stream that will be read in each XREADGROUP of all streams
	// Example: if there are 2 streams and MaxMsgCountPerRead=10, then each XREADGROUP may read upto a total of 20 messages.
	// Same as COUNT option in XREADGROUP(https://redis.io/topics/streams-intro). Defaults to 10
	// +optional
	MaxMsgCountPerRead int32 `json:"maxMsgCountPerRead,omitempty" protobuf:"varint,5,opt,name=maxMsgCountPerRead"`
	// ConsumerGroup refers to the Redis stream consumer group that will be
	// created on all redis streams. Messages are read through this group. Defaults to 'argo-events-cg'
	// +optional
	ConsumerGroup string `json:"consumerGroup,omitempty" protobuf:"bytes,6,opt,name=consumerGroup"`
	// TLS configuration for the redis client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,7,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,8,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,9,opt,name=filter"`
	// Username required for ACL style authentication if any.
	// +optional
	Username string `json:"username,omitempty" protobuf:"bytes,10,opt,name=username"`
}

// NSQEventSource describes the event source for NSQ PubSub
// More info at https://godoc.org/github.com/nsqio/go-nsq
type NSQEventSource struct {
	// HostAddress is the address of the host for NSQ lookup
	HostAddress string `json:"hostAddress" protobuf:"bytes,1,opt,name=hostAddress"`
	// Topic to subscribe to.
	Topic string `json:"topic" protobuf:"bytes,2,opt,name=topic"`
	// Channel used for subscription
	Channel string `json:"channel" protobuf:"bytes,3,opt,name=channel"`
	// Backoff holds parameters applied to connection.
	// +optional
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,4,opt,name=connectionBackoff"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,5,opt,name=jsonBody"`
	// TLS configuration for the nsq client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,7,rep,name=metadata"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,8,opt,name=filter"`
}

// PulsarEventSource describes the event source for Apache Pulsar
type PulsarEventSource struct {
	// Name of the topics to subscribe to.
	// +required
	Topics []string `json:"topics" protobuf:"bytes,1,rep,name=topics"`
	// Type of the subscription.
	// Only "exclusive" and "shared" is supported.
	// Defaults to exclusive.
	// +optional
	Type string `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
	// Configure the service URL for the Pulsar service.
	// +required
	URL string `json:"url" protobuf:"bytes,3,name=url"`
	// Trusted TLS certificate secret.
	// +optional
	TLSTrustCertsSecret *corev1.SecretKeySelector `json:"tlsTrustCertsSecret,omitempty" protobuf:"bytes,4,opt,name=tlsTrustCertsSecret"`
	// Whether the Pulsar client accept untrusted TLS certificate from broker.
	// +optional
	TLSAllowInsecureConnection bool `json:"tlsAllowInsecureConnection,omitempty" protobuf:"bytes,5,opt,name=tlsAllowInsecureConnection"`
	// Whether the Pulsar client verify the validity of the host name from broker.
	// +optional
	TLSValidateHostname bool `json:"tlsValidateHostname,omitempty" protobuf:"bytes,6,opt,name=tlsValidateHostname"`
	// TLS configuration for the pulsar client.
	// +optional
	TLS *apicommon.TLSConfig `json:"tls,omitempty" protobuf:"bytes,7,opt,name=tls"`
	// Backoff holds parameters applied to connection.
	// +optional
	ConnectionBackoff *apicommon.Backoff `json:"connectionBackoff,omitempty" protobuf:"bytes,8,opt,name=connectionBackoff"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"bytes,9,opt,name=jsonBody"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,10,rep,name=metadata"`
	// Authentication token for the pulsar client.
	// +optional
	AuthTokenSecret *corev1.SecretKeySelector `json:"authTokenSecret,omitempty" protobuf:"bytes,11,opt,name=authTokenSecret"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,12,opt,name=filter"`
}

// GenericEventSource refers to a generic event source. It can be used to implement a custom event source.
type GenericEventSource struct {
	// URL of the gRPC server that implements the event source.
	URL string `json:"url" protobuf:"bytes,1,name=url"`
	// Config is the event source configuration
	Config string `json:"config" protobuf:"bytes,2,name=config"`
	// Insecure determines the type of connection.
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,3,opt,name=insecure"`
	// JSONBody specifies that all event body payload coming from this
	// source will be JSON
	// +optional
	JSONBody bool `json:"jsonBody,omitempty" protobuf:"varint,4,opt,name=jsonBody"`
	// Metadata holds the user defined metadata which will passed along the event payload.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,5,rep,name=metadata"`
	// AuthSecret holds a secret selector that contains a bearer token for authentication
	// +optional
	AuthSecret *corev1.SecretKeySelector `json:"authSecret,omitempty" protobuf:"bytes,6,opt,name=authSecret"`
	// Filter
	// +optional
	Filter *EventSourceFilter `json:"filter,omitempty" protobuf:"bytes,7,opt,name=filter"`
}

const (
	// EventSourceConditionSourcesProvided has the status True when the EventSource
	// has its event source provided.
	EventSourceConditionSourcesProvided apicommon.ConditionType = "SourcesProvided"
	// EventSourceConditionDeployed has the status True when the EventSource
	// has its Deployment created.
	EventSourceConditionDeployed apicommon.ConditionType = "Deployed"
)

// EventSourceStatus holds the status of the event-source resource
type EventSourceStatus struct {
	apicommon.Status `json:",inline" protobuf:"bytes,1,opt,name=status"`
}

// InitConditions sets conditions to Unknown state.
func (es *EventSourceStatus) InitConditions() {
	es.InitializeConditions(EventSourceConditionSourcesProvided, EventSourceConditionDeployed)
}

// MarkSourcesProvided set the eventsource has valid sources spec provided.
func (es *EventSourceStatus) MarkSourcesProvided() {
	es.MarkTrue(EventSourceConditionSourcesProvided)
}

// MarkSourcesNotProvided the eventsource has invalid sources spec provided.
func (es *EventSourceStatus) MarkSourcesNotProvided(reason, message string) {
	es.MarkFalse(EventSourceConditionSourcesProvided, reason, message)
}

// MarkDeployed set the eventsource has been deployed.
func (es *EventSourceStatus) MarkDeployed() {
	es.MarkTrue(EventSourceConditionDeployed)
}

// MarkDeployFailed set the eventsource deploy failed
func (es *EventSourceStatus) MarkDeployFailed(reason, message string) {
	es.MarkFalse(EventSourceConditionDeployed, reason, message)
}
