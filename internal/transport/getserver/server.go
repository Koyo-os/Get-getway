package getserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/Koyo-os/get-getway/internal/core"
	"github.com/Koyo-os/get-getway/pkg/api/protobuf/get"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GetServer struct {
	get.UnimplementedGetServiceServer
	logger *logger.Logger
	server *grpc.Server
	core   *core.GetServiceCore
}

// Get implements get.GetServiceServer.
func (g *GetServer) Get(ctx context.Context, req *get.GetRequest) (*get.GetResponse, error) {
	if req == nil {
		return nil, errors.New("req is nil")
	}

	switch req.GetType {
	case "single":
		resp, err := g.core.RouteGetRequest(req.Payload, req.EntityName)
		if err != nil {
			return &get.GetResponse{
				Entities: nil,
				Response: &get.Response{
					Ok:    false,
					Error: err.Error(),
				},
			}, fmt.Errorf("error get "+req.EntityName+": %v", err)
		}

		return &get.GetResponse{
			Entities: []string{string(resp)},
			Response: &get.Response{
				Ok:    true,
				Error: "",
			},
		}, nil
	case "many":
		resp, err := g.core.RouteGetMoreRequest(req.Payload, req.EntityName)
		if err != nil {
			return &get.GetResponse{
				Response: &get.Response{
					Error: err.Error(),
					Ok:    false,
				},

				Entities: nil,
			}, err
		}

		return &get.GetResponse{
			Entities: resp,
			Response: &get.Response{
				Ok:    true,
				Error: "",
			},
		}, nil
	default:
		message := "unknown get type"

		return &get.GetResponse{
			Response: &get.Response{
				Ok:    false,
				Error: message,
			},
			Entities: nil,
		}, errors.New(message)
	}
}

// GetRealTime implements get.GetServiceServer.
func (g *GetServer) GetRealTime(req *get.GetRealTimeRequest, resp grpc.ServerStreamingServer[get.GetResponse]) error {
	if req == nil{
		return errors.New("request is nil")
	}
	
	respChan, err := g.core.RouteGetRealTimeRequest(req.Payload, req.EntityName)
	if err != nil{
		return fmt.Errorf("error route real-time request: %v", err)
	}

	for r := range respChan{
		if err = resp.Send(&get.GetResponse{
			Response: &get.Response{
				Ok: true,
				Error: "",
			},
			Entities: []string{r},
		});err != nil{
			g.logger.Error("failed send entity", zap.Error(err))

			continue
		}
	}

	return nil
}

// mustEmbedUnimplementedGetServiceServer implements get.GetServiceServer.

func NewGetServer(core *core.GetServiceCore) *GetServer {
	server := grpc.NewServer()

	get.RegisterGetServiceServer(server, &GetServer{})

	return &GetServer{
		logger: logger.Get(),
		server: server,
		core: core,
	}
}
