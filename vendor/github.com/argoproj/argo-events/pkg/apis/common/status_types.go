package common

import (
	"reflect"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType is a valid value of Condition.Type
type ConditionType string

const (
	// ConditionReady indicates the resource is ready.
	ConditionReady ConditionType = "Ready"
)

// Condition contains details about resource state
type Condition struct {
	// Condition type.
	// +required
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Condition status, True, False or Unknown.
	// +required
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// Unique, this should be a short, machine understandable string that gives the reason
	// for condition's last transition. For example, "ImageNotFound"
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// IsTrue tells if the condition is True
func (c *Condition) IsTrue() bool {
	if c == nil {
		return false
	}
	return c.Status == corev1.ConditionTrue
}

// IsFalse tells if the condition is False
func (c *Condition) IsFalse() bool {
	if c == nil {
		return false
	}
	return c.Status == corev1.ConditionFalse
}

// IsUnknown tells if the condition is Unknown
func (c *Condition) IsUnknown() bool {
	if c == nil {
		return true
	}
	return c.Status == corev1.ConditionUnknown
}

// GetReason returns as Reason
func (c *Condition) GetReason() string {
	if c == nil {
		return ""
	}
	return c.Reason
}

// GetMessage returns a Message
func (c *Condition) GetMessage() string {
	if c == nil {
		return ""
	}
	return c.Message
}

// Status is a common structure which can be used for Status field.
type Status struct {
	// Conditions are the latest available observations of a resource's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// InitializeConditions initializes the contions to Unknown
func (s *Status) InitializeConditions(conditionTypes ...ConditionType) {
	for _, t := range conditionTypes {
		c := Condition{
			Type:   t,
			Status: corev1.ConditionUnknown,
		}
		s.SetCondition(c)
	}
}

// SetCondition sets a condition
func (s *Status) SetCondition(condition Condition) {
	var conditions []Condition
	for _, c := range s.Conditions {
		if c.Type != condition.Type {
			conditions = append(conditions, c)
		} else {
			condition.LastTransitionTime = c.LastTransitionTime
			if reflect.DeepEqual(&condition, &c) {
				return
			}
		}
	}
	condition.LastTransitionTime = metav1.NewTime(time.Now())
	conditions = append(conditions, condition)
	// Sort for easy read
	sort.Slice(conditions, func(i, j int) bool { return conditions[i].Type < conditions[j].Type })
	s.Conditions = conditions
}

func (s *Status) markTypeStatus(t ConditionType, status corev1.ConditionStatus, reason, message string) {
	s.SetCondition(Condition{
		Type:    t,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
}

// MarkTrue sets the status of t to true
func (s *Status) MarkTrue(t ConditionType) {
	s.markTypeStatus(t, corev1.ConditionTrue, "", "")
}

// MarkTrueWithReason sets the status of t to true with reason
func (s *Status) MarkTrueWithReason(t ConditionType, reason, message string) {
	s.markTypeStatus(t, corev1.ConditionTrue, reason, message)
}

// MarkFalse sets the status of t to fasle
func (s *Status) MarkFalse(t ConditionType, reason, message string) {
	s.markTypeStatus(t, corev1.ConditionFalse, reason, message)
}

// MarkUnknown sets the status of t to unknown
func (s *Status) MarkUnknown(t ConditionType, reason, message string) {
	s.markTypeStatus(t, corev1.ConditionUnknown, reason, message)
}

// GetCondition returns the condition of a condtion type
func (s *Status) GetCondition(t ConditionType) *Condition {
	for _, c := range s.Conditions {
		if c.Type == t {
			return &c
		}
	}
	return nil
}

// IsReady returns true when all the conditions are true
func (s *Status) IsReady() bool {
	if len(s.Conditions) == 0 {
		return false
	}
	for _, c := range s.Conditions {
		if !c.IsTrue() {
			return false
		}
	}
	return true
}
