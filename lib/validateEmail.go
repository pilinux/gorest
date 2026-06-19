package lib

import (
	"context"
	"net"
	"regexp"
	"strings"
	"time"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// EmailMXLookupTimeout bounds the DNS MX lookup performed by ValidateEmail.
// Without a bound the lookup blocks for the OS resolver's default timeout,
// turning email validation into a latency/DoS amplification surface when an
// attacker submits domains that are slow to resolve.
var EmailMXLookupTimeout = 5 * time.Second

// ValidateEmailFormat checks if the email provided passes the required
// structure and length test.
//
// Credit: Edd Turtle
func ValidateEmailFormat(e string) bool {
	if len(e) < 3 || len(e) > 254 {
		return false
	}

	return emailRegex.MatchString(e)
}

// ValidateEmail checks if the email provided passes the required structure
// and length test. It also checks the domain has a valid MX record.
//
// Note: this function performs a live, blocking DNS query. The lookup is bounded
// by EmailMXLookupTimeout, but it still depends on network state, so its result
// is non-deterministic.
func ValidateEmail(e string) bool {
	if !ValidateEmailFormat(e) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), EmailMXLookupTimeout)
	defer cancel()

	_, domain, ok := strings.Cut(e, "@")
	if !ok {
		return false
	}
	mx, err := net.DefaultResolver.LookupMX(ctx, domain)
	// Note: RFC 5321 §5.1 allows a sender to fall back to the domain's A/AAAA
	// record when no MX record exists (implicit MX). gorest intentionally does
	// not support this fallback: modern mail servers rely almost exclusively on
	// MX records, so a domain that accepts mail only via an implicit A/AAAA
	// record is highly suspicious today and is treated as invalid here.
	if err != nil || len(mx) == 0 {
		return false
	}

	// RFC 7505
	// https://www.rfc-editor.org/rfc/rfc7505.html
	if mx[0].Host == "." {
		return false
	}

	return true
}
