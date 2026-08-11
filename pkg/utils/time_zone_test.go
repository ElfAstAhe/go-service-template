package utils

import (
	"testing"
	_ "time/tzdata"
)

func TestValidateTimeZone(t *testing.T) {
	// Проверяем, что списки не пустые после init()
	if len(SortedTimeZones) == 0 {
		t.Fatal("SortedTimeZones is empty after initialization")
	}

	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{
			name:    "Valid standard time zone",
			tz:      "Europe/Moscow",
			wantErr: false,
		},
		{
			name:    "Valid US time zone",
			tz:      "America/New_York",
			wantErr: false,
		},
		{
			name:    "Valid UTC",
			tz:      "UTC",
			wantErr: false,
		},
		{
			name:    "Invalid time zone name",
			tz:      "Europe/Moskva",
			wantErr: true,
		},
		{
			name:    "Empty string",
			tz:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeZone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeZone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeZoneSortedListAreActuallySorted(t *testing.T) {
	// Проверяем корректность сортировки таймзон
	for i := 1; i < len(SortedTimeZones); i++ {
		if SortedTimeZones[i-1] > SortedTimeZones[i] {
			t.Errorf("SortedTimeZones is not sorted at index %d: %s > %s", i, SortedTimeZones[i-1], SortedTimeZones[i])
		}
	}
}
