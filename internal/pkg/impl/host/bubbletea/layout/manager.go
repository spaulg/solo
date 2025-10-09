package layout

import (
	"charm.land/lipgloss/v2"
)

type Manager struct {
	layoutDirection Direction
	layoutSpecs     []Spec
}

type Direction int

type Type int

const (
	HorizontalLayoutDirection Direction = 1
	VerticalLayoutDirection   Direction = 2

	PercentageLayoutType Type = 1
	FixedLayoutType      Type = 2
	FillLayoutType       Type = 3
)

func NewLayoutManager(layoutDirection Direction, layoutSpecs []Spec) *Manager {
	var fillSpec *Spec
	for i, spec := range layoutSpecs {
		if spec.Type != FillLayoutType {
			continue
		}

		if fillSpec != nil {
			// If multiple fill specs are found, silently override
			// extra ones to a percentage layout with 0 size
			// This avoids having to raise and return an error throughout the
			// model constructor hierarchy
			layoutSpecs[i].Type = PercentageLayoutType
			layoutSpecs[i].Size = 0
		}

		fillSpec = &spec
	}

	if fillSpec == nil {
		layoutSpecs = append(layoutSpecs, NewFillLayoutSpec(lipgloss.NewStyle()))
	}

	return &Manager{
		layoutDirection: layoutDirection,
		layoutSpecs:     layoutSpecs,
	}
}

func (t *Manager) Render(sourceStrings ...string) string {
	renderedStrings := make([]string, len(sourceStrings))

	for i, str := range sourceStrings {
		renderedStrings[i] = t.layoutSpecs[i].Style.Render(str)
	}

	if t.layoutDirection == HorizontalLayoutDirection {
		return lipgloss.JoinHorizontal(lipgloss.Left, renderedStrings...)
	} else {
		return lipgloss.JoinVertical(lipgloss.Top, renderedStrings...)
	}
}

func (t *Manager) CalculateDimensions(w int, h int) []*Dimension {
	var axisDimensions []*int
	dimensions := make([]*Dimension, 0)

	// Calculate axis sizes
	if t.layoutDirection == HorizontalLayoutDirection {
		axisDimensions = t.calculateAxisDimensions(w)
	} else {
		axisDimensions = t.calculateAxisDimensions(h)
	}

	// Reformat the axis sizes in to an
	// array of dimension structures
	for i, axisDimension := range axisDimensions {
		var dimension *Dimension

		if t.layoutDirection == HorizontalLayoutDirection {
			dimension = &Dimension{
				Style:  t.layoutSpecs[i].Style,
				Width:  *axisDimension,
				Height: h,
			}
		} else {
			dimension = &Dimension{
				Style:  t.layoutSpecs[i].Style,
				Width:  w,
				Height: *axisDimension,
			}
		}

		dimensions = append(dimensions, dimension)
	}

	return dimensions
}

func (t *Manager) calculateAxisDimensions(axis int) []*int {
	runningSize := 0
	dimensions := make([]*int, 0)

	var fillDimension *int
	var fillLayoutSpec *Spec

	for _, layoutSpec := range t.layoutSpecs {
		calculatedSize := 0

		if layoutSpec.Type != FillLayoutType {
			calculatedSize = layoutSpec.CalculateBorderBoxSize(axis, runningSize)
			runningSize += calculatedSize
		}

		dimension := &calculatedSize
		dimensions = append(dimensions, dimension)

		if layoutSpec.Type == FillLayoutType {
			fillDimension = dimension
			fillLayoutSpec = &layoutSpec
		}
	}

	if fillLayoutSpec != nil && fillDimension != nil {
		*fillDimension = fillLayoutSpec.CalculateBorderBoxSize(axis, runningSize)
	}

	return dimensions
}
