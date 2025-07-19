package scheduler

import "context"

// JobHandler represents a scheduled job that can be executed
type JobHandler interface {
	Name() string
	// Execute executes the job with the given parameters
	Execute(ctx context.Context, params ...any) error
	// ValidateParams validates that the parameters are correct for this job
	ValidateParams(params []any) error
}

// JobFunc is a helper type that allows plain functions to be used as Jobs
type JobFunc func(ctx context.Context, params []any) error

func (f JobFunc) Run(ctx context.Context, params []any) error {
	return f(ctx, params)
}

func (f JobFunc) Name() string {
	return "anonymous"
}
