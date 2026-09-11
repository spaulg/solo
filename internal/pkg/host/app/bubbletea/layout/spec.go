package layout

import (
	"math"

	"charm.land/lipgloss/v2"
)

type Spec struct {
	Type  Type
	Style lipgloss.Style
	Size  int
}

func NewPercentageLayoutSpec(size int, style lipgloss.Style) Spec {
	return Spec{
		Type:  PercentageLayoutType,
		Size:  size,
		Style: style,
	}
}

func NewFixedLayoutSpec(size int, style lipgloss.Style) Spec {
	return Spec{
		Type:  FixedLayoutType,
		Size:  size,
		Style: style,
	}
}

func NewFillLayoutSpec(style lipgloss.Style) Spec {
	return Spec{
		Type:  FillLayoutType,
		Style: style,
	}
}

func (t Spec) CalculateBorderBoxSize(totalSize int, usedSize int) int {
	switch t.Type {
	case PercentageLayoutType:
		return int(math.Round(float64(totalSize) * (float64(t.Size) / 100)))
	case FixedLayoutType:
		return t.Size
	case FillLayoutType:
		return totalSize - usedSize
	default:
		return 0
	}
}
