package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Koyo-os/get-getway/internal/config"
	"github.com/Koyo-os/get-getway/internal/entity"
	"github.com/Koyo-os/get-getway/pkg/clients/answer"
	"github.com/Koyo-os/get-getway/pkg/clients/form"
	"github.com/Koyo-os/get-getway/pkg/clients/poll"
	"github.com/Koyo-os/get-getway/pkg/clients/vote"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	GetServiceCore struct {
		logger *logger.Logger
		cfg    *config.Config
		form   *form.FormClient
		answer *answer.AnswerClient
		vote   *vote.VoteClient
		poll   *poll.PollClient
	}

	Connection struct {
		formClient   *form.FormClient
		answerClient *answer.AnswerClient
		voteClient   *vote.VoteClient
		pollClient   *poll.PollClient
	}
)

var ErrUnknownEntityType error = errors.New("unknown entity type")

func ConnectToServices(
	urls map[string]string,
) (*Connection, error) {
	conn := &Connection{}

	for entity, url := range urls {
		cc, err := grpc.NewClient(url)
		if err != nil {
			return nil, err
		}

		switch entity {
		case "answer":
			conn.answerClient = answer.NewAnswerClient(cc, 20*time.Second)
		case "poll":
			conn.formClient = form.NewFormClient(cc, 20*time.Second)
		case "vote":
			conn.voteClient = vote.NewVoteClient(cc, 20*time.Second)
		case "form":
			conn.formClient = form.NewFormClient(cc, 20*time.Second)
		default:
			continue
		}
	}

	return conn, nil
}

func NewGetServiceCore(
	conn *Connection,
) *GetServiceCore {
	return &GetServiceCore{
		form:   conn.formClient,
		answer: conn.answerClient,
		vote:   conn.voteClient,
		poll:   conn.pollClient,
		logger: logger.Get(),
		cfg:    config.NewConfig(),
	}
}

func (g *GetServiceCore) RouteGetRequest(payload string, entityType string) ([]byte, error) {
	request := &entity.GetOne{}

	if len(payload) <= 1 {
		return nil, errors.New("empty payload")
	}

	if err := sonic.Unmarshal([]byte(payload), request); err != nil {
		return nil, fmt.Errorf("failed to get id from payload: %v", err)
	}

	switch entityType {
	case "poll":
		poll, err := g.poll.Get(request.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get poll: %v", err)
		}

		return sonic.Marshal(poll)
	case "answer":
		answer, err := g.answer.Get(request.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get answer: %v", err)
		}

		return sonic.Marshal(answer)
	case "form":
		form, err := g.form.Get(request.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get form: %v", err)
		}

		return sonic.Marshal(form)
	default:
		g.logger.Warn("unknown entity type", zap.String("type", entityType))
		return nil, ErrUnknownEntityType
	}
}

func (g *GetServiceCore) RouteGetMoreRequest(payload, entityType string) ([]string, error) {
	if len(payload) == 0 {
		return nil, errors.New("payload is nil")
	}

	req := &entity.GetMore{}

	if err := sonic.Unmarshal([]byte(payload), req); err != nil {
		return nil, fmt.Errorf("failed unmarshal payload: %v", err)
	}

	if len(req.Key) == 0 && len(req.Value) == 0 {
		return nil, errors.New("value or key is nil")
	}

	switch entityType {
	case "poll":
		resp, err := g.poll.GetMore(req.Key, req.Value)
		if err != nil {
			return nil, fmt.Errorf("error get more for poll: %v", err)
		}

		entities := make([]string, len(resp))

		for i, r := range resp {
			entities[i], err = sonic.MarshalString(r)
			if err != nil {
				continue
			}
		}

		return entities, nil
	case "answer":
		resp, err := g.answer.GetMore(req.Key, req.Value)
		if err != nil {
			return nil, fmt.Errorf("error get more for poll: %v", err)
		}

		entities := make([]string, len(resp))

		for i, r := range resp {
			entities[i], err = sonic.MarshalString(r)
			if err != nil {
				continue
			}
		}

		return entities, nil
	case "form":
		resp, err := g.answer.GetMore(req.Key, req.Value)
		if err != nil {
			return nil, fmt.Errorf("error get more forms: %v", err)
		}

		entities := make([]string, len(resp))

		for i, r := range resp {
			entities[i], err = sonic.MarshalString(r)
			if err != nil {
				continue
			}
		}

		return entities, nil

	default:
		return nil, ErrUnknownEntityType
	}
}

func (g *GetServiceCore) RouteGetRealTimeRequest(payload, entityType string) (chan string, error) {
	respChan := make(chan string, 1)

	if len(payload) == 0 {
		return respChan, errors.New("payload is nil")
	}

	req := &entity.GetOne{}

	if err := sonic.Unmarshal([]byte(payload), req); err != nil {
		return respChan, fmt.Errorf("failed unmarshal payload: %v", err)
	}

	switch entityType {
	case "vote":
		resp, err := g.vote.Get(context.Background(), req.ID)
		if err != nil {
			return respChan, fmt.Errorf("error get realtime for vote: %v", err)
		}

		go func() {
			for r := range resp {
				voteJson, err := sonic.Marshal(&r)
				if err != nil {
					g.logger.Error("error marshal vote", zap.Error(err))

					continue
				}

				respChan <- string(voteJson)
			}
		}()

		return respChan, nil
	default:
		return respChan, ErrUnknownEntityType
	}
}

func (g *GetServiceCore) RouteEvent(event *entity.Event) error {
	firstPart := strings.Split(event.Type, ".")[0]

	switch firstPart {
	case "answer":
		return route(g.answer, event.Type, string(event.Payload))
	case "poll":
		return route(g.poll, event.Type, string(event.Payload))
	case "form":
		return route(g.form, event.Type, string(event.Payload))
	case "vote":
		return route(g.vote, event.Type, string(event.Payload))
	default:
		return ErrUnknownEntityType
	}
}
