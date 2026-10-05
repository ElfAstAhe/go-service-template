package worker

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewBaseSchedulerDispatcherOptions_Defaults(t *testing.T) {
	opts := NewBaseSchedulerDispatcherOptions[string]()

	assert.Equal(t, DefaultSchedulerDispatcherPoolWorkerCount, opts.WorkerCount)
	assert.Equal(t, DefaultSchedulerDispatcherPoolDataCapacity, opts.DataCapacity)
	assert.Equal(t, DefaultSchedulerDispatcherPoolCompleteProcess, opts.CompleteProcess)
	assert.Equal(t, DefaultSchedulerDispatcherSchedulerStartInterval, opts.StartInterval)
	assert.Equal(t, DefaultSchedulerDispatcherSchedulerScheduleInterval, opts.ScheduleInterval)
	assert.Equal(t, DefaultSchedulerDispatcherStopTimeout, opts.StopTimeout)
	assert.Empty(t, opts.Name)
	assert.Nil(t, opts.Logger)
	assert.Nil(t, opts.JobHandler)
	assert.Nil(t, opts.DataProvider)
}

func TestBaseSchedulerDispatcherOptions_FluentAPI(t *testing.T) {
	mockLogger := mocks.NewMockLogger(t)
	var dummyHandler JobHandler[string] = func(ctx context.Context, workerIndex int, data string) error { return nil }
	var dummyProvider DispatcherDataProvider[string] = func(ctx context.Context, eventTime time.Time) ([]string, error) { return nil, nil }

	opts := NewBaseSchedulerDispatcherOptions[string]()

	WithSchedulerDispatcherName[string]("custom-dispatcher")(opts)
	WithSchedulerDispatcherPoolWorkerCount[string](5)(opts)
	WithSchedulerDispatcherPoolDataCapacity[string](256)(opts)
	WithSchedulerDispatcherPoolCompleteProcess[string](false)(opts)
	WithSchedulerDispatcherPoolJobHandler[string](dummyHandler)(opts)
	WithSchedulerDispatcherSchedulerStartInterval[string](time.Millisecond * 25)(opts)
	WithSchedulerDispatcherSchedulerScheduleInterval[string](time.Second * 10)(opts)
	WithSchedulerDispatcherStopTimeout[string](time.Second * 4)(opts)
	WithSchedulerDispatcherDataProvider[string](dummyProvider)(opts)
	WithSchedulerDispatcherLogger[string](mockLogger)(opts)

	assert.Equal(t, "custom-dispatcher", opts.Name)
	assert.Equal(t, 5, opts.WorkerCount)
	assert.Equal(t, 256, opts.DataCapacity)
	assert.False(t, opts.CompleteProcess)
	assert.Equal(t, time.Millisecond*25, opts.StartInterval)
	assert.Equal(t, time.Second*10, opts.ScheduleInterval)
	assert.Equal(t, time.Second*4, opts.StopTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.NotNil(t, opts.JobHandler)
	assert.NotNil(t, opts.DataProvider)
}

func TestBaseSchedulerDispatcherOptions_Validate(t *testing.T) {
	var validHandler JobHandler[string] = func(ctx context.Context, workerIndex int, data string) error { return nil }
	var validProvider DispatcherDataProvider[string] = func(ctx context.Context, eventTime time.Time) ([]string, error) { return nil, nil }

	tests := []struct {
		name    string
		modify  func(*testing.T, *BaseSchedulerDispatcherOptions[string])
		wantErr bool
	}{
		{
			name: "valid options",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: false,
		},
		{
			name: "empty name",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "   "
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "invalid worker count",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.WorkerCount = -1
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "invalid data capacity",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.DataCapacity = 0
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "invalid stop timeout",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.StopTimeout = 0
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "nil job handler",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.Logger = mockLogger
				o.JobHandler = nil
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "nil logger",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				o.Name = "valid-dispatcher"
				o.Logger = nil
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "nil data provider",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = nil
			},
			wantErr: true,
		},
		{
			name: "invalid start interval",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.StartInterval = 0
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
		{
			name: "invalid schedule interval",
			modify: func(t *testing.T, o *BaseSchedulerDispatcherOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-dispatcher"
				o.ScheduleInterval = -time.Millisecond
				o.Logger = mockLogger
				o.JobHandler = validHandler
				o.DataProvider = validProvider
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewBaseSchedulerDispatcherOptions[string]()
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
