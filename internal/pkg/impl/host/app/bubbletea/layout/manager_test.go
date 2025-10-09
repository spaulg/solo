package layout

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLayoutManagerTestSuite(t *testing.T) {
	suite.Run(t, new(LayoutManagerTestSuite))
}

type LayoutManagerTestSuite struct {
	suite.Suite
}

func (t *LayoutManagerTestSuite) TestNoFillLayoutDefaults() {
	layoutManager := NewLayoutManager(HorizontalLayoutDirection, []Spec{
		{
			Type: PercentageLayoutType,
			Size: 25,
		},
	})

	dimensions := layoutManager.CalculateDimensions(251, 71)
	t.Assert().Len(dimensions, 2)

	t.Equal(63, dimensions[0].Width)
	t.Equal(71, dimensions[0].Height)

	t.Equal(188, dimensions[1].Width)
	t.Equal(71, dimensions[1].Height)
}

func (t *LayoutManagerTestSuite) TestMultipleFillReplaces() {
	layoutManager := NewLayoutManager(HorizontalLayoutDirection, []Spec{
		{
			Type: FillLayoutType,
		},
		{
			Type: FillLayoutType,
		},
	})

	dimensions := layoutManager.CalculateDimensions(251, 71)
	t.Assert().Len(dimensions, 2)

	t.Equal(251, dimensions[0].Width)
	t.Equal(71, dimensions[0].Height)

	t.Equal(0, dimensions[1].Width)
	t.Equal(71, dimensions[1].Height)
}

func (t *LayoutManagerTestSuite) TestHorizontalLayout() {
	layoutManager := NewLayoutManager(HorizontalLayoutDirection, []Spec{
		{
			Type: PercentageLayoutType,
			Size: 25,
		},
		{
			Type: PercentageLayoutType,
			Size: 25,
		},
		{
			Type: FillLayoutType,
		},
	})

	dimensions := layoutManager.CalculateDimensions(251, 71)
	t.Assert().Len(dimensions, 3)

	t.Equal(63, dimensions[0].Width)
	t.Equal(71, dimensions[0].Height)

	t.Equal(63, dimensions[1].Width)
	t.Equal(71, dimensions[1].Height)

	t.Equal(125, dimensions[2].Width)
	t.Equal(71, dimensions[2].Height)
}

func (t *LayoutManagerTestSuite) TestVerticalLayout() {
	layoutManager := NewLayoutManager(VerticalLayoutDirection, []Spec{
		{
			Type: PercentageLayoutType,
			Size: 25,
		},
		{
			Type: PercentageLayoutType,
			Size: 25,
		},
		{
			Type: FillLayoutType,
		},
	})

	dimensions := layoutManager.CalculateDimensions(251, 71)
	t.Assert().Len(dimensions, 3)

	t.Equal(251, dimensions[0].Width)
	t.Equal(18, dimensions[0].Height)

	t.Equal(251, dimensions[1].Width)
	t.Equal(18, dimensions[1].Height)

	t.Equal(251, dimensions[2].Width)
	t.Equal(35, dimensions[2].Height)
}

// Percentage only
// Percentage with min & max
// Percentage with style
// Percentage with min & max, with style
// Percentage + fill

// Fixed only
// Fixed with style
// Fixed + fill

// Percentage + fixed

// Unhappy paths
//  no fill type - 1 fill type is always required
//  more than one fill type - there can be only one
