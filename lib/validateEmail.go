package lib

import (
	"net"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// ValidateEmail checks if the email provided passes the required structure
// and length test. It also checks the domain has a valid MX record.
//
// Credit: Edd Turtle
func ValidateEmail(e string) bool {
	if len(e) < 3 || len(e) > 254 {
		return false
	}

	if !emailRegex.MatchString(e) {
		return false
	}

	parts := strings.Split(e, "@")
	mx, err := net.LookupMX(parts[1])
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
