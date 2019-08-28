package health

import "context"

type Check struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Code     int32  `json:"code"`
	Endpoint string `json:"endpoint"`
	Checked  int64  `json:"checked"`
	Duration string `json:"duration"`
}

type SVC interface {
	Create(ctx context.Context, endpoint string) (Check, error)
}

type service struct {
}

var _ SVC = (*service)(nil)

func (s *service) Create(ctx context.Context, endpoint string) (Check, error) {
	return Check{}, nil
}
