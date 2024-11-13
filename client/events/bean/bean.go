package bean

import "github.com/devtron-labs/devtron/util/event"

type Provider struct {
	Destination util.Channel `json:"dest"`
	Rule        string       `json:"rule"`
	ConfigId    int          `json:"configId"`
	Recipient   string       `json:"recipient"`
}

const SES_CONFIG_TYPE = "ses"
const SMTP_CONFIG_TYPE = "smtp"
