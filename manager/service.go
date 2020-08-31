package manager

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type service struct {
	l log.Logger
	p Publisher
}

// Service implements business logic
type Service interface {
	Subscribe(ctx context.Context, streamer, authorization string) error
}

var (
	ErrTest = errors.New("test")
)

// NewService returns an account Service with all of the expected middlewares wired in.
func NewService(
	l log.Logger,
	p Publisher,
) Service {
	s := service{}

	s.l = l
	s.p = p

	return &s
}

func (s *service) Subscribe(ctx context.Context, streamer, authorization string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", streamer), nil)
	req.Header.Add("Authorization", authorization)
	client.Do(req)

	resp, err := http.Get(fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", streamer))
	if err != nil {
		fmt.Println("Error getting user id...")
		level.Error(s.l).Log("method", "Subscribe", "action", "http.Get", "error", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		level.Error(s.l).Log("method", "Subscribe", "action", "ioutil.ReadAll", "error", err.Error())
	}

	userID := string(body)
	fmt.Println(userID)

	var (
		mode        = "subscribe"
		topic       = fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&to_id=%s", userID)
		accessToken = fmt.Sprintf("%v", authorization)
	)

	s.p.PublishSubscribeToStreamerMsg(ctx, SubscribeMsg{mode: mode, topic: topic, secret: accessToken})

	return nil
}
