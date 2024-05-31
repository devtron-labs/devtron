/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

type EventType int

const (
	Trigger                   EventType = 1
	Success                   EventType = 2
	Fail                      EventType = 3
	Approval                  EventType = 4
	ConfigApproval            EventType = 5
	Blocked                   EventType = 6
	ArtifactPromotionApproval EventType = 7
	ImageScanning             EventType = 8
)
const ScoopNotification EventType = 9

var RulesSupportedEvents = []int{int(ImageScanning)}

type PipelineType string

const CI PipelineType = "CI"
const CD PipelineType = "CD"
const PRE_CD PipelineType = "PRE-CD"
const POST_CD PipelineType = "POST-CD"

type Level string

type Channel string

const (
	Slack   Channel = "slack"
	SES     Channel = "ses"
	SMTP    Channel = "smtp"
	Webhook Channel = "webhook"
)
const PANIC = "panic"

type UpdateType string

const (
	UpdateEvents     UpdateType = "events"
	UpdateRecipients UpdateType = "recipients"
	UpdateRules      UpdateType = "rules"
)
