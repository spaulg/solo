package layout

import (
	"charm.land/lipgloss/v2"
)

type Dimension struct {
	Style  lipgloss.Style
	Width  int
	Height int
}

func (t Dimension) ContentBoxWidth() int {
	return t.Width -
		t.Style.GetBorderLeftSize() - t.Style.GetBorderRightSize() -
		t.Style.GetPaddingLeft() - t.Style.GetPaddingRight()
}

func (t Dimension) ContentBoxHeight() int {
	return t.Height -
		t.Style.GetBorderTopSize() - t.Style.GetBorderBottomSize() -
		t.Style.GetPaddingTop() - t.Style.GetPaddingBottom()
}
