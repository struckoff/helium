package helium

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/struckoff/helium/module"
	"github.com/struckoff/helium/service"
)

type errService struct {
	start bool
	stop  bool

	*atomic.Error
}

func (e *errService) Start(_ context.Context) error {
	if !e.start {
		return nil
	}

	return errTest
}

func (e *errService) Stop(context.Context) {
	if e.stop {
		e.Store(errTest)
	}
}

func (e errService) Name() string { return "errService" }

func TestDefaultApp(t *testing.T) {
	t.Run("create new helium with default application", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			module.New(viper.New),
			module.New(zap.NewNop),
			module.New(func() context.Context { return ctx }),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		cancel()

		require.NoError(t, h.Run())
	})

	t.Run("default application with start err", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		h, err := New(&Settings{},
			DefaultApp,
			module.New(viper.New),
			module.New(zap.NewNop),
			module.New(func() context.Context { return ctx }),
			module.New(func() service.Service { return &errService{start: true} }, dig.Group("services")),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		require.EqualError(t, h.Run(), errTest.Error())

		cancel()
	})

	t.Run("default application with stop err", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		svc := &errService{stop: true, Error: atomic.NewError(nil)}

		h, err := New(&Settings{},
			DefaultApp,
			module.New(viper.New),
			module.New(func() context.Context { return ctx }),
			module.New(func() *zap.Logger { return zaptest.NewLogger(t) }),
			module.New(func() service.Service { return svc }, dig.Group("services")),
		)

		require.NotNil(t, h)
		require.NoError(t, err)

		cancel()
		require.NoError(t, h.Run())
		require.EqualError(t, svc.Load(), errTest.Error())
	})
}
