package lib

import (
	"net"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// InitRateLimiter - initialize the rate limiter instance
func InitRateLimiter(formattedRateLimit, trustedPlatform string) (*limiter.Limiter, error) {
	if formattedRateLimit == "" {
		return nil, nil
	}

	rate, err := limiter.NewRateFromFormatted(formattedRateLimit)
	if err != nil {
		return nil, err
	}

	// use an in-memory store with a goroutine which clears expired keys
	store := memory.NewStore()

	// custom IPv6 mask
	ipv6Mask := net.CIDRMask(64, 128)

	options := []limiter.Option{
		limiter.WithIPv6Mask(ipv6Mask),
	}
	if trustedPlatform != "" {
		options = append(options, limiter.WithClientIPHeader(trustedPlatform))
	}

	// create the limiter instance
	limiterInstance := limiter.New(
		store,
		rate,
		options...,
	)

	return limiterInstance, nil
}
