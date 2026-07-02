package notification

import (
	"context"
	"sync"

	"github.com/devops/agent/internal/domain/agent"
)

type MultiNotificationRouter struct {
	Sinks []agent.NotificationSink
}

func NewMultiNotificationRouter(sinks ...agent.NotificationSink) *MultiNotificationRouter {
	return &MultiNotificationRouter{
		Sinks: sinks,
	}
}

func (m *MultiNotificationRouter) Emit(ctx context.Context, topic string, payload map[string]interface{}) error {
	var wg sync.WaitGroup
	for _, sink := range m.Sinks {
		wg.Add(1)
		go func(s agent.NotificationSink) {
			defer wg.Done()
			_ = s.Emit(ctx, topic, payload)
		}(sink)
	}
	wg.Wait()
	return nil
}

var _ agent.NotificationSink = (*MultiNotificationRouter)(nil)
