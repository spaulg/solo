package progress

import "github.com/stretchr/testify/suite"

type ProgressEntityTypeNameTestSuite struct {
	suite.Suite
}

func (t *ProgressEntityTypeNameTestSuite) TestEntityTypeNameFromString() {
	t.Equal(Volume, EntityTypeNameFromString("Volume"))
	t.Equal(Network, EntityTypeNameFromString("Network"))
	t.Equal(Container, EntityTypeNameFromString("Container"))
	t.Equal(Image, EntityTypeNameFromString("Image"))

	t.Equal(UnknownEntityType, EntityTypeNameFromString("qwerty"))
}

func (t *ProgressEntityTypeNameTestSuite) TestString() {
	t.Equal("Volume", Volume.String())
	t.Equal("Network", Network.String())
	t.Equal("Container", Container.String())
	t.Equal("Image", Image.String())

	t.Equal("Unknown", UnknownEntityType.String())
}
