package worker

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewBaseSchedulerOptions_Defaults(t *testing.T) {
	opts := NewBaseSchedulerOptions()

	assert.Equal(t, defaultSchedulerStartInterval, opts.StartInterval)
	assert.Equal(t, defaultSchedulerScheduleInterval, opts.ScheduleInterval)
	assert.Equal(t, defaultSchedulerStopTimeout, opts.StopTimeout)
	assert.Empty(t, opts.Name)
	assert.Nil(t, opts.Logger)
	assert.Nil(t, opts.TimerDispatcher)
}

func TestBaseSchedulerOptions_FluentAPI(t *testing.T) {
	mockLogger := mocks.NewMockLogger(t)
	var dummyDispatcher TimerDispatcher = func(ctx context.Context, eventTime time.Time) error { return nil }

	opts := NewBaseSchedulerOptions()

	WithSchedulerName("custom-scheduler")(opts)
	WithSchedulerStartInterval(time.Millisecond * 15)(opts)
	WithSchedulerScheduleInterval(time.Millisecond * 45)(opts)
	WithSchedulerStopTimeout(time.Second * 3)(opts)
	WithSchedulerLogger(mockLogger)(opts)
	WithSchedulerTimerDispatcher(dummyDispatcher)(opts)

	assert.Equal(t, "custom-scheduler", opts.Name)
	assert.Equal(t, time.Millisecond*15, opts.StartInterval)
	assert.Equal(t, time.Millisecond*45, opts.ScheduleInterval)
	assert.Equal(t, time.Second*3, opts.StopTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.NotNil(t, opts.TimerDispatcher)
}

func TestBaseSchedulerOptions_Validate(t *testing.T) {
	var validDispatcher TimerDispatcher = func(ctx context.Context, eventTime time.Time) error { return nil }

	tests := []struct {
		name    string
		modify  func(*testing.T, *BaseSchedulerOptions)
		wantErr bool
	}{
		{
			name: "valid options",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-scheduler"
				o.Logger = mockLogger
				o.TimerDispatcher = validDispatcher
			},
			wantErr: false,
		},
		{
			name: "empty name",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "   "
				o.Logger = mockLogger
				o.TimerDispatcher = validDispatcher
			},
			wantErr: true,
		},
		{
			name: "invalid start interval",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-scheduler"
				o.StartInterval = 0
				o.Logger = mockLogger
				o.TimerDispatcher = validDispatcher
			},
			wantErr: true,
		},
		{
			name: "invalid schedule interval",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-scheduler"
				o.ScheduleInterval = -time.Second
				o.Logger = mockLogger
				o.TimerDispatcher = validDispatcher
			},
			wantErr: true,
		},
		{
			name: "invalid stop timeout",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-scheduler"
				o.StopTimeout = 0
				o.Logger = mockLogger
				o.TimerDispatcher = validDispatcher
			},
			wantErr: true,
		},
		{
			name: "nil timer dispatcher",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-scheduler"
				o.Logger = mockLogger
				o.TimerDispatcher = nil
			},
			wantErr: true,
		},
		{
			name: "nil logger",
			modify: func(t *testing.T, o *BaseSchedulerOptions) {
				o.Name = "valid-scheduler"
				o.Logger = nil
				o.TimerDispatcher = validDispatcher
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewBaseSchedulerOptions()
			tt.modify(t, opts)
			err := opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
