package migration

import (
	"testing"
)

func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name        string
		databaseDSN string
		wantErr     bool
	}{
		{
			name:        "empty DSN should return error",
			databaseDSN: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RunMigrations(tt.databaseDSN); (err != nil) != tt.wantErr {
				t.Errorf("RunMigrations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
