package notifications

import "context"

// PushFunc adapts an application push provider without coupling storage to it.
type PushFunc func(context.Context, Notification) error

func (f PushFunc) Send(ctx context.Context, notification Notification) error {
	return f(ctx, notification)
}
