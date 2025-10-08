package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	t.Parallel()
	s := NewStore()
	assert.NotNil(t, s)
}

func TestStoreSet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		key  string
		val  string
		ttl  *time.Duration
	}{
		{"no ttl", "k1", "v1", nil},
		{"with ttl future", "k2", "v2", func() *time.Duration { d := 100 * time.Millisecond; return &d }()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			s.Set(tc.key, tc.val, tc.ttl)
			got, ok := s.Get(tc.key)
			assert.True(t, ok)
			assert.Equal(t, tc.val, got)
		})
	}
}

func TestStoreGet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		setupFunc func(s *Store)
		key       string
		wantVal   string
		wantOK    bool
	}{
		{
			name:      "missing key",
			setupFunc: func(s *Store) {},
			key:       "k0",
			wantVal:   "",
			wantOK:    false,
		},
		{
			name:      "present no ttl",
			setupFunc: func(s *Store) { s.Set("k1", "v1", nil) },
			key:       "k1",
			wantVal:   "v1",
			wantOK:    true,
		},
		{
			name: "expired ttl",
			setupFunc: func(s *Store) {
				d := -1 * time.Second // expired in the past
				s.Set("k2", "v2", &d)
			},
			key:     "k2",
			wantVal: "",
			wantOK:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			tc.setupFunc(s)
			got, ok := s.Get(tc.key)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.wantVal, got)
		})
	}
}

func TestStoreIncr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		setupFunc func(s *Store)
		key       string
		wantVal   int
		wantErr   bool
	}{
		{
			name:      "new key starts at 1",
			setupFunc: func(s *Store) {},
			key:       "k3",
			wantVal:   1,
			wantErr:   false,
		},
		{
			name:      "existing integer increments",
			setupFunc: func(s *Store) { s.Set("k4", "41", nil) },
			key:       "k4",
			wantVal:   42,
			wantErr:   false,
		},
		{
			name:      "non-integer value returns error",
			setupFunc: func(s *Store) { s.Set("k5", "abc", nil) },
			key:       "k5",
			wantVal:   0,
			wantErr:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := NewStore()
			tc.setupFunc(s)
			got, err := s.Incr(tc.key)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantVal, got)
		})
	}
}
