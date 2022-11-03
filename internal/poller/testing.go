package poller

import (
	"context"
	"github.com/vlad-marlo/gophermart/internal/store"
	"testing"
)

type testPoller struct {
	store store.Storage
}

// Register ...
func (s *testPoller) Register(ctx context.Context, user, num int) error {
	return s.store.Order().Register(ctx, user, num)
}
func (s *testPoller) Close() {}

func TestPoller(t *testing.T, store store.Storage) OrderPoller {
	t.Helper()
	return &testPoller{store: store}
}
