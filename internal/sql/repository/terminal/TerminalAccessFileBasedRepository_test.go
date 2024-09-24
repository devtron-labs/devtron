package terminal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTerminalAccessFileBasedRepository(t *testing.T) {
	t.SkipNow()
	t.Run("access templates", func(t *testing.T) {
		fileBasedRepository := NewTerminalAccessFileBasedRepository(nil)
		templates, err := fileBasedRepository.FetchAllTemplates()
		assert.Nil(t, err)
		for _, template := range templates {
			fmt.Println(template)
		}
	})
}
