package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/stretchr/testify/suite"
)

type WorkflowGuardFactoryTestSuite struct {
	suite.Suite
}

func (t *WorkflowGuardFactoryTestSuite) TestBuild() {
	soloCtx := &context.CliContext{}

	workflowNames := []workflowcommon.WorkflowName{workflowcommon.FirstPreStart, workflowcommon.PreStart, workflowcommon.PostStart}
	containerNames := []string{"container1", "container2"}

	workflowGuardFactory := NewWorkflowGuardFactory(soloCtx)
	workflowGuard := workflowGuardFactory.Build(workflowNames, containerNames)

	t.NotNil(workflowGuard)
}
