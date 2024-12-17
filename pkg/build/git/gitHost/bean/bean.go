package bean

type GitHostRequest struct {
	Id              int    `json:"id,omitempty" validate:"number"`
	Name            string `json:"name,omitempty" validate:"required"`
	Active          bool   `json:"active"`
	WebhookUrl      string `json:"webhookUrl"`
	WebhookSecret   string `json:"webhookSecret"`
	EventTypeHeader string `json:"eventTypeHeader"`
	SecretHeader    string `json:"secretHeader"`
	SecretValidator string `json:"secretValidator"`
	UserId          int32  `json:"-"`
}

const (
	HostUrlPlaceHolder = "{HOST_URL_PLACEHOLDER}"
)
