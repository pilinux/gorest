package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// firewall package-level variables
var (
	parsedOnce sync.Once
	ipNets     []*net.IPNet
	ipListMap  map[string]bool
	ipCIDR     bool
)

// Firewall - whitelist/blacklist IPs
func Firewall(listType string, ipList string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse the IP list only once
		parsedOnce.Do(func() {
			parseIPList(listType, ipList)
		})

		// get the real client IP
		clientNetIP := net.ParseIP(c.ClientIP())
		if clientNetIP == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "IP invalid")
			return
		}
		clientIP := clientNetIP.String()

		if !strings.Contains(ipList, "*") {
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

		if strings.Contains(ipList, "*") {
			if listType == "blacklist" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, "IP blocked")
				return
			}
		}

		c.Next()
	}
}

// helper function to parse the IP list and CIDR notations
func parseIPList(listType, ipList string) {
	ipListMap = make(map[string]bool)

	// split the list by comma and trim spaces
	ipListSlice := strings.Split(ipList, ",")
	for _, ip := range ipListSlice {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if strings.Contains(ip, "/") {
			// parse CIDR notations
			_, ipNet, err := net.ParseCIDR(ip)
			if err == nil {
				ipNets = append(ipNets, ipNet)
			}
		} else {
			ipListMap[ip] = true
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
		if strings.Contains(validIPs, "*") {
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
		if strings.Contains(validIPs, "*") {
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
}

// ResetFirewallState - helper function to reset firewall package-level variables
func ResetFirewallState() {
	parsedOnce = sync.Once{}
	ipNets = nil
	ipListMap = nil
	ipCIDR = false
}
