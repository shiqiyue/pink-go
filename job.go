package client

import "context"

type Job interface {
	Target() string
	// execute job
	Execute(ctx context.Context, param string) (string, error)
}
