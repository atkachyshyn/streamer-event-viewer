package collector

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
)

type service struct {
	l log.Logger
}

// Service implements business logic
type Service interface {
	Subscribe(ctx context.Context, mode, topic, secret string) error
	ListenToEvents(ctx context.Context) error
}

var (
	ErrTest = errors.New("test")
)

const (
	leaseSeconds = "864000"
)

// NewService returns an account Service with all of the expected middlewares wired in.
func NewService(
	l log.Logger,
) Service {
	s := service{}

	s.l = l

	return &s
}

func (s *service) Subscribe(ctx context.Context, mode, topic, secret string) error {
	http.PostForm("https://api.twitch.tv/helix/webhooks/hub",
		url.Values{
			"hub.callback":      {"http://localhost:3000/callback"},
			"hub.mode":          {mode},
			"hub.topic":         {topic},
			"hub.lease_seconds": {leaseSeconds},
			"hub.secret":        {secret},
		})

	return nil
}

func (s *service) ListenToEvents(ctx context.Context) error {
	return nil
}
