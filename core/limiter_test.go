package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	cosmos := "cosmos"
	osmosis := "osmosis"
	chains := []ChainId{cosmos, osmosis}
	limitPeriod := int64(1)

	tests := map[string]struct {
		init   func() *Limiter
		expect func(t *testing.T, l *Limiter)
	}{
		"empty limiter always returns true": {
			init: func() *Limiter {
				return NewLimiter(chains, limitPeriod)
			},
			expect: func(t *testing.T, l *Limiter) {
				assert.Equal(t, l.IsAllowed(cosmos, "address1"), true)
				assert.Equal(t, l.IsAllowed(osmosis, "address1"), true)
			},
		},
		"user can request token to different chains without limit": {
			init: func() *Limiter {
				l := NewLimiter(chains, limitPeriod)
				l.AddRequest(cosmos, "address1")
				return l
			},
			expect: func(t *testing.T, l *Limiter) {
				assert.Equal(t, false, l.IsAllowed(cosmos, "address1"))
				assert.Equal(t, true, l.IsAllowed(osmosis, "address1"))
			},
		},
		"user cannot request token within limit period": {
			init: func() *Limiter {
				l := NewLimiter(chains, limitPeriod)
				l.AddRequest(cosmos, "address1")
				return l
			},
			expect: func(t *testing.T, l *Limiter) {
				assert.Equal(t, false, l.IsAllowed(cosmos, "address1"))
			},
		},
		"user can request token after limit period": {
			init: func() *Limiter {
				l := NewLimiter(chains, limitPeriod)
				l.AddRequest(cosmos, "address1")
				l.AddRequest(osmosis, "address1")
				return l
			},
			expect: func(t *testing.T, l *Limiter) {
				time.Sleep(time.Duration(limitPeriod+1) * time.Second)
				assert.Equal(t, true, l.IsAllowed(cosmos, "address1"))
				assert.Equal(t, true, l.IsAllowed(osmosis, "address1"))
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			l := test.init()
			test.expect(t, l)
		})
	}
}
