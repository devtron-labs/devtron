package bean

import (
	"time"
)

type TeamRequest struct {
	Id        int       `json:"id,omitempty" validate:"number"`
	Name      string    `json:"name,omitempty" validate:"required"`
	Active    bool      `json:"active"`
	UserId    int32     `json:"-"`
	CreatedOn time.Time `json:"-"`
}

type TeamBean struct {
	Id   int    `json:"id"`
	Name string `json:"name,notnull"`
}
