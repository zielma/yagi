package jobs

import "github.com/zielma/yagi/internal/scheduler"

func init() {
	scheduler.RegisterJob("syncBudgets", syncBudgets)
}
