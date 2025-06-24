package listener

import (
	"context"

	"github.com/Koyo-os/get-getway/internal/entity"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"go.uber.org/zap"
)

type (
	Router interface {
		RouteEvent(*entity.Event) error
	}

	Listener struct {
		InputChan chan entity.Event
		logger    *logger.Logger
		Router    Router
	}
)

func NewListener(input chan entity.Event, router Router) *Listener {
	return &Listener{
		InputChan: input,
		Router:    router,
		logger:    logger.Get(),
	}
}

func (l *Listener) Listen(ctx context.Context) {
	for {
		select {
		case event := <-l.InputChan:
			if err := l.Router.RouteEvent(&event); err != nil {
				l.logger.Error("error route event",
					zap.String("event_id", event.ID),
					zap.Error(err))
			}
		case <-ctx.Done():
			l.logger.Info("listener stopped")
			return
		}
	}
}
