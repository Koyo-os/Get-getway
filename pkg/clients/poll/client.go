package poll

import (
	"context"
	"errors"
	"time"

	pollpb "github.com/Koyo-os/get-getway/pkg/api/protobuf/poll"
	"github.com/Koyo-os/get-getway/pkg/clients/poll/entity"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PollClient struct {
	client  pollpb.PollServiceClient
	logger  *logger.Logger
	timeout time.Duration
}

func NewPollClient(cc grpc.ClientConnInterface, timeout time.Duration) *PollClient {
	return &PollClient{
		client:  pollpb.NewPollServiceClient(cc),
		logger:  logger.Get(),
		timeout: timeout,
	}
}

func (p *PollClient) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), p.timeout)
}

func (p *PollClient) Create(poll *entity.Poll) error {
	ctx, cancel := p.context()
	defer cancel()

	resp, err := p.client.Create(ctx, &pollpb.RequestCreate{
		Form: poll.ToProtobuf(),
	})
	if err != nil {
		p.logger.Error("error create poll",
			zap.String("poll_id", poll.ID.String()),
			zap.Error(err))
		return err
	}

	if !resp.Ok {
		p.logger.Error("error from response", zap.String("error", resp.Error))
		return errors.New(resp.Error)
	}

	return nil
}

func (p *PollClient) Get(id string) (*entity.Poll, error) {
	ctx, cancel := p.context()
	defer cancel()

	resp, err := p.client.Get(ctx, &pollpb.RequestGet{ID: id})
	if err != nil {
		p.logger.Error("error get poll", zap.String("poll_id", id), zap.Error(err))
		return nil, err
	}
	if !resp.Response.Ok {
		p.logger.Error("error from response", zap.String("error", resp.Response.Error))
		return nil, errors.New(resp.Response.Error)
	}

	pollEntity, err := entity.ToEntityPoll(resp.Poll)
	if err != nil {
		p.logger.Error("error converting poll to entity", zap.Error(err))
		return nil, err
	}
	return pollEntity, nil
}

func (p *PollClient) GetMore(key, value string) ([]*entity.Poll, error) {
	ctx, cancel := p.context()
	defer cancel()

	resp, err := p.client.GetMore(ctx, &pollpb.RequestGetMore{
		Key:   key,
		Value: value,
	})
	if err != nil {
		p.logger.Error("error get more polls", zap.Error(err))
		return nil, err
	}
	if !resp.Response.Ok {
		p.logger.Error("error from response", zap.String("error", resp.Response.Error))
		return nil, errors.New(resp.Response.Error)
	}

	result := make([]*entity.Poll, 0, len(resp.Forms))
	for _, pbPoll := range resp.Forms {
		pollEntity, err := entity.ToEntityPoll(pbPoll)
		if err != nil {
			p.logger.Warn("error converting poll to entity", zap.Error(err))
			continue
		}
		result = append(result, pollEntity)
	}
	return result, nil
}

func (p *PollClient) Update(id, key, value string) error {
	ctx, cancel := p.context()
	defer cancel()

	resp, err := p.client.Update(ctx, &pollpb.RequestUpdate{
		ID:    id,
		Key:   key,
		Value: value,
	})
	if err != nil {
		p.logger.Error("error update poll", zap.String("poll_id", id), zap.Error(err))
		return err
	}
	if !resp.Ok {
		p.logger.Error("error from response", zap.String("error", resp.Error))
		return errors.New(resp.Error)
	}
	return nil
}

func (p *PollClient) Delete(id string) error {
	ctx, cancel := p.context()
	defer cancel()

	resp, err := p.client.Delete(ctx, &pollpb.RequestDelete{ID: id})
	if err != nil {
		p.logger.Error("error delete poll", zap.String("poll_id", id), zap.Error(err))
		return err
	}
	if !resp.Ok {
		p.logger.Error("error from response", zap.String("error", resp.Error))
		return errors.New(resp.Error)
	}
	return nil
}
