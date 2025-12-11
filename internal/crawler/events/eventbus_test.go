package events_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/domain"
	crawlerMock "github.com/jonesrussell/gocrawl/testutils/mocks/crawler"
	loggerMock "github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEventBus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockLog := loggerMock.NewMockInterface(ctrl)
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	bus := events.NewEventBus(mockLog)

	t.Run("NewEventBus", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, bus)
	})

	t.Run("Subscribe", func(t *testing.T) {
		t.Parallel()
		testCtrl := gomock.NewController(t)
		t.Cleanup(testCtrl.Finish)

		testMockLog := loggerMock.NewMockInterface(testCtrl)
		testMockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		testBus := events.NewEventBus(testMockLog)
		handler := crawlerMock.NewMockEventHandler(testCtrl)

		var receivedArticle *domain.Article
		handler.EXPECT().HandleArticle(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, article *domain.Article) {
				receivedArticle = article
			}).
			Return(nil).
			Times(1)

		testBus.Subscribe(handler)
		article := &domain.Article{Title: "Test Article"}
		err := testBus.PublishArticle(context.Background(), article)
		require.NoError(t, err)
		assert.Equal(t, article, receivedArticle)
	})

	t.Run("PublishError", func(t *testing.T) {
		t.Parallel()
		testCtrl := gomock.NewController(t)
		t.Cleanup(testCtrl.Finish)

		testMockLog := loggerMock.NewMockInterface(testCtrl)
		testMockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		testBus := events.NewEventBus(testMockLog)
		handler := crawlerMock.NewMockEventHandler(testCtrl)

		var receivedErr error
		handler.EXPECT().HandleError(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, err error) {
				receivedErr = err
			}).
			Return(nil).
			Times(1)

		testBus.Subscribe(handler)
		testErr := errors.New("test error")
		testBus.PublishError(context.Background(), testErr)
		assert.Equal(t, testErr, receivedErr)
	})

	t.Run("PublishStart", func(t *testing.T) {
		t.Parallel()
		testCtrl := gomock.NewController(t)
		t.Cleanup(testCtrl.Finish)

		testMockLog := loggerMock.NewMockInterface(testCtrl)
		testMockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		testBus := events.NewEventBus(testMockLog)
		handler := crawlerMock.NewMockEventHandler(testCtrl)

		handler.EXPECT().HandleStart(gomock.Any()).Return(nil).Times(1)

		testBus.Subscribe(handler)
		err := testBus.PublishStart(context.Background())
		require.NoError(t, err)
	})

	t.Run("PublishStop", func(t *testing.T) {
		t.Parallel()
		testCtrl := gomock.NewController(t)
		t.Cleanup(testCtrl.Finish)

		testMockLog := loggerMock.NewMockInterface(testCtrl)
		testMockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		testBus := events.NewEventBus(testMockLog)
		handler := crawlerMock.NewMockEventHandler(testCtrl)

		handler.EXPECT().HandleStop(gomock.Any()).Return(nil).Times(1)

		testBus.Subscribe(handler)
		err := testBus.PublishStop(context.Background())
		require.NoError(t, err)
	})
}

// TestEventBus_ConcurrentSubscribe tests concurrent Subscribe calls
func TestEventBus_ConcurrentSubscribe(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockLog := loggerMock.NewMockInterface(ctrl)
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	bus := events.NewEventBus(mockLog)

	// Create handlers with expectations before concurrent access
	handlers := make([]*crawlerMock.MockEventHandler, 100)
	for i := range 100 {
		handlers[i] = crawlerMock.NewMockEventHandler(ctrl)
		handlers[i].EXPECT().HandleError(gomock.Any(), gomock.Any()).AnyTimes()
		handlers[i].EXPECT().HandleArticle(gomock.Any(), gomock.Any()).AnyTimes()
		handlers[i].EXPECT().HandleStart(gomock.Any()).AnyTimes()
		handlers[i].EXPECT().HandleStop(gomock.Any()).AnyTimes()
	}

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bus.Subscribe(handlers[idx])
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 100, bus.HandlerCount())
}

// TestEventBus_ConcurrentPublish tests concurrent Publish calls
func TestEventBus_ConcurrentPublish(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockLog := loggerMock.NewMockInterface(ctrl)
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	bus := events.NewEventBus(mockLog)
	handler := crawlerMock.NewMockEventHandler(ctrl)
	handler.EXPECT().HandleError(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	bus.Subscribe(handler)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.PublishError(context.Background(), errors.New("test"))
		}()
	}

	wg.Wait()
	// Should not panic or race
}

// TestEventBus_ConcurrentSubscribeAndPublish tests Subscribe and Publish concurrently
func TestEventBus_ConcurrentSubscribeAndPublish(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockLog := loggerMock.NewMockInterface(ctrl)
	mockLog.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	bus := events.NewEventBus(mockLog)

	// Create handlers with expectations before concurrent access
	handlers := make([]*crawlerMock.MockEventHandler, 100)
	for i := range 100 {
		handlers[i] = crawlerMock.NewMockEventHandler(ctrl)
		handlers[i].EXPECT().HandleError(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		handlers[i].EXPECT().HandleArticle(gomock.Any(), gomock.Any()).AnyTimes()
		handlers[i].EXPECT().HandleStart(gomock.Any()).AnyTimes()
		handlers[i].EXPECT().HandleStop(gomock.Any()).AnyTimes()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Publishers
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				select {
				case <-ctx.Done():
					return
				default:
					bus.PublishError(context.Background(), errors.New("test"))
				}
			}
		}()
	}

	// Subscribers
	handlerIdx := 0
	var mu sync.Mutex
	const subscriberCount = 10
	const subscribePerSubscriber = 10
	for i := range subscriberCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range subscribePerSubscriber {
				select {
				case <-ctx.Done():
					return
				default:
					mu.Lock()
					idx := handlerIdx
					handlerIdx++
					mu.Unlock()
					bus.Subscribe(handlers[idx])
				}
				_ = j
			}
		}()
		_ = i
	}

	wg.Wait()
}
