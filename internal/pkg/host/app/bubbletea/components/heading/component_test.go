package heading

import (
	"testing"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
)

func TestHeadingComponentTestSuite(t *testing.T) {
	suite.Run(t, new(HeadingComponentTestSuite))
}

type HeadingComponentTestSuite struct {
	suite.Suite
}

func (t *HeadingComponentTestSuite) TestRender() {
	model := NewComponent("Test Heading")

	updated, _ := model.Update(messages.ComponentSizeMsg{
		Height: 1,
		Width:  300,
	})

	view := updated.(Component).View().Content

	teatest.RequireEqualOutput(t.T(), []byte(view))
}
