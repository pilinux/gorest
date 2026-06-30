package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Firewall applies whitelist/blacklist IP filtering.
func Firewall(listType string, ipList string) gin.HandlerFunc {
	// parse the IP list once at construction time so each middleware instance
	// keeps its own state instead of sharing package-level globals
	ipNets, ipListMap, ipCIDR, wildcard := parseIPList(listType, ipList)

	return func(c *gin.Context) {
		// get the real client IP
		clientNetIP := net.ParseIP(c.ClientIP())
		if clientNetIP == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "IP invalid")
			return
		}
		clientIP := clientNetIP.String()

		if !wildcard {
			if listType == "whitelist" {
				var allowIP bool
				if len(ipListMap) > 0 {
					if _, ok := ipListMap[clientIP]; ok {
						allowIP = true
					}
				}
				if !allowIP && ipCIDR {
					for _, ipNet := range ipNets {
						if ipNet.Contains(clientNetIP) {
							allowIP = true
							break
						}
					}
				}
				if !allowIP {
					c.AbortWithStatusJSON(http.StatusUnauthorized, "IP blocked")
					return
				}
			}

			if listType == "blacklist" {
				var blockIP bool
				if len(ipListMap) > 0 {
					if _, ok := ipListMap[clientIP]; ok {
						blockIP = true
					}
				}
				if !blockIP && ipCIDR {
					for _, ipNet := range ipNets {
						if ipNet.Contains(clientNetIP) {
							blockIP = true
							break
						}
					}
				}
				if blockIP {
					c.AbortWithStatusJSON(http.StatusUnauthorized, "IP blocked")
					return
				}
			}
		}

		if wildcard {
			if listType == "blacklist" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, "IP blocked")
				return
			}
		}

		c.Next()
	}
}

// helper function to parse the IP list and CIDR notations
func parseIPList(listType, ipList string) (ipNets []*net.IPNet, ipListMap map[string]bool, ipCIDR bool, wildcard bool) {
	ipListMap = make(map[string]bool)

	// split the list by comma and trim spaces
	ipSeq := strings.SplitSeq(ipList, ",")
	for ip := range ipSeq {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		// only an exact "*" entry means match-all
		if ip == "*" {
			wildcard = true
			continue
		}
		if strings.Contains(ip, "/") {
			// parse CIDR notations
			_, ipNet, err := net.ParseCIDR(ip)
			if err == nil {
				ipNets = append(ipNets, ipNet)
			}
		} else {
			// normalize the exact IP to its canonical form so the map key
			// matches the client IP lookup key (net.IP.String()); this makes
			// matching independent of the case/form used in the config (e.g.
			// IPv6 "2400:CB00::1" vs "2400:cb00::1"). Invalid entries are skipped.
			if normalized := net.ParseIP(ip); normalized != nil {
				ipListMap[normalized.String()] = true
			}
		}
	}

	// if any CIDR notations were found, set ipCIDR to true
	if len(ipNets) > 0 {
		ipCIDR = true
	}

	var validIPs string
	var validCIDRs string
	for ip := range ipListMap {
		validIPs += ip + ", "
	}
	for _, ipNet := range ipNets {
		validCIDRs += ipNet.String() + ", "
	}
	// remove the trailing comma and space
	validIPs = strings.TrimSuffix(validIPs, ", ")
	validCIDRs = strings.TrimSuffix(validCIDRs, ", ")

	fmt.Println("application firewall initialized")
	if listType == "whitelist" {
		if wildcard {
			fmt.Println("whitelisted IPs: *")
		} else {
			if len(validIPs) > 0 {
				fmt.Println("whitelisted IPs:", validIPs)
			}
			if len(validCIDRs) > 0 {
				fmt.Println("whitelisted CIDRs:", validCIDRs)
			}
		}
	}
	if listType == "blacklist" {
		if wildcard {
			fmt.Println("blacklisted IPs: *")
		} else {
			if len(validIPs) > 0 {
				fmt.Println("blacklisted IPs:", validIPs)
			}
			if len(validCIDRs) > 0 {
				fmt.Println("blacklisted CIDRs:", validCIDRs)
			}
		}
	}

	return ipNets, ipListMap, ipCIDR, wildcard
}
