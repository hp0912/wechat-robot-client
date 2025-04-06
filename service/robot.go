package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type RobotService struct {
	ctx context.Context
}

func NewRobotService(ctx context.Context) *RobotService {
	return &RobotService{
		ctx: ctx,
	}
}

func (r *RobotService) IsRunning() bool {
	client := resty.New()
	resp, err := client.R().Get(fmt.Sprintf("%s/IsRunning", vars.RobotRuntime.Doman()))
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("Error checking if robot is running: %v\nhttp code: %d", err, resp.StatusCode())
		return false
	}
	return resp.String() == "OK"
}
