package worker

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/worker"
	"github.com/stretchr/testify/assert"
)

// Константы дефолтов для внутренней защиты рантайм-компонента (независимые от пакета config)
const (
	defaultSchedulerStartInterval    time.Duration = time.Second
	defaultSchedulerScheduleInterval time.Duration = time.Second * 5
	defaultSchedulerStopTimeout      time.Duration = time.Second * 5
)

func TestNewJanitorOptions_Defaults(t *testing.T) {
	opts := NewJanitorOptions[string, int]()

	assert.NotNil(t, opts.BaseSchedulerOptions)
	assert.Equal(t, defaultSchedulerStartInterval, opts.StartInterval)
	assert.Equal(t, defaultSchedulerScheduleInterval, opts.ScheduleInterval)
	assert.Equal(t, defaultSchedulerStopTimeout, opts.StopTimeout)
	assert.Nil(t, opts.Cache)
}

func TestJanitorOptions_FluentAPI(t *testing.T) {
	mockCache := mocks.NewMockCache[string, int](t)
	mockLogger := mocks2.NewMockLogger(t)

	opts := NewJanitorOptions[string, int]()

	WithJanitorName[string, int]("cache-janitor")(opts)
	WithJanitorStartInterval[string, int](15 * time.Millisecond)(opts)
	WithJanitorScheduleInterval[string, int](45 * time.Millisecond)(opts)
	WithJanitorStopTimeout[string, int](time.Second * 2)(opts)
	WithJanitorLogger[string, int](mockLogger)(opts)
	WithJanitorCache[string, int](mockCache)(opts)

	assert.Equal(t, "cache-janitor", opts.Name)
	assert.Equal(t, 15*time.Millisecond, opts.StartInterval)
	assert.Equal(t, 45*time.Millisecond, opts.ScheduleInterval)
	assert.Equal(t, time.Second*2, opts.StopTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.Equal(t, mockCache, opts.Cache)
}

func TestJanitorOptions_Validate(t *testing.T) {
	mockCache := mocks.NewMockCache[string, int](t)
	mockLogger := mocks2.NewMockLogger(t)

	var dummyDispatcher worker.TimerDispatcher = func(ctx context.Context, t time.Time) error { return nil }

	tests := []struct {
		name    string
		modify  func(*JanitorOptions[string, int])
		wantErr bool
	}{
		{
			name: "valid janitor options",
			modify: func(o *JanitorOptions[string, int]) {
				o.Name = "valid-janitor"
				o.Logger = mockLogger
				o.TimerDispatcher = dummyDispatcher
				o.Cache = mockCache
			},
			wantErr: false,
		},
		{
			name: "failed base validation empty name",
			modify: func(o *JanitorOptions[string, int]) {
				o.Name = ""
				o.Logger = mockLogger
				o.TimerDispatcher = dummyDispatcher
				o.Cache = mockCache
			},
			wantErr: true,
		},
		{
			name: "failed specific validation nil cache",
			modify: func(o *JanitorOptions[string, int]) {
				o.Name = "valid-janitor"
				o.Logger = mockLogger
				o.TimerDispatcher = dummyDispatcher
				o.Cache = nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewJanitorOptions[string, int]()
			tt.modify(opts)
			err := opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
