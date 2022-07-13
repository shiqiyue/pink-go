package client

import "context"

type Extension interface {
	// 执行任务前
	Before(ctx context.Context, task Job) context.Context

	// 执行任务后
	After(ctx context.Context, task Job) context.Context
}
