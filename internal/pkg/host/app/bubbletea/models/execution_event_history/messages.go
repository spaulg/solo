package execution_event_history

import (
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/components/table"
)

type ExecutionEventHistoryDataLoaded struct {
	rows table.Rows
}
