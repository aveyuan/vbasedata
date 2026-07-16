package vbasedata

import "testing"

func TestNewRedis_RejectsEmptyAddresses(t *testing.T) {
	cases := []*RedisConfig{
		{},
		{Addr: []string{""}},
		{Addr: []string{"  "}},
	}
	for _, cfg := range cases {
		client, closeFn, err := NewRedis(cfg, discardLogger())
		if err == nil {
			t.Errorf("NewRedis(%+v) returned nil error", cfg)
		}
		if client != nil || closeFn != nil {
			t.Errorf("NewRedis(%+v) returned resources on validation failure", cfg)
		}
	}
}
