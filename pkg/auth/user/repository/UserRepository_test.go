package repository

import (
	"testing"
)

func TestFilterActiveUserIds(t *testing.T) {
	users := []UserModel{
		{Id: 1, Active: true},
		{Id: 2, Active: false},
		{Id: 3, Active: true},
	}

	active := FilterActiveUserIds(users)
	if len(active) != 2 {
		t.Fatalf("expected 2 active users, got %d", len(active))
	}
	if active[0] != 1 || active[1] != 3 {
		t.Errorf("unexpected active user ids: %v", active)
	}
}
