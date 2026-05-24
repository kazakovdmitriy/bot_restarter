package main

import "testing"

func TestIsAllowed(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		allowedIDs []int64
		want       bool
	}{
		{
			name:       "user is in list",
			userID:     123,
			allowedIDs: []int64{100, 123, 200},
			want:       true,
		},
		{
			name:       "user is not in list",
			userID:     999,
			allowedIDs: []int64{100, 200},
			want:       false,
		},
		{
			name:       "empty list",
			userID:     123,
			allowedIDs: []int64{},
			want:       false,
		},
		{
			name:       "nil list",
			userID:     123,
			allowedIDs: nil,
			want:       false,
		},
		{
			name:       "single element match",
			userID:     42,
			allowedIDs: []int64{42},
			want:       true,
		},
		{
			name:       "single element no match",
			userID:     42,
			allowedIDs: []int64{99},
			want:       false,
		},
		{
			name:       "negative user ID match",
			userID:     -1,
			allowedIDs: []int64{-1, 0, 1},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAllowed(tt.userID, tt.allowedIDs)
			if got != tt.want {
				t.Errorf("IsAllowed(%d, %v) = %v, want %v", tt.userID, tt.allowedIDs, got, tt.want)
			}
		})
	}
}
