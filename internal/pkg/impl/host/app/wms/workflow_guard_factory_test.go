package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

func TestWorkflowGuardFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowGuardFactoryTestSuite))
}

type WorkflowGuardFactoryTestSuite struct {
	suite.Suite
}

func (t *WorkflowGuardFactoryTestSuite) TestBuild() {
	soloCtx := &context.CliContext{}

	workflowNames := []workflowcommon.WorkflowName{workflowcommon.FirstPreStartContainer, workflowcommon.PreStartContainer, workflowcommon.PostStartContainer}
	containerNames := []string{"container1", "container2"}

	workflowGuardFactory := NewWorkflowGuardFactory(soloCtx)
	workflowGuard := workflowGuardFactory.Build(workflowNames, containerNames)

	t.NotNil(workflowGuard)
}
