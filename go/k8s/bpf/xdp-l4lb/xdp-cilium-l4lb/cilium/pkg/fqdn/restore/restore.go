package restore

import "regexp"

// DNSRules contains IP-based DNS rules for a set of ports (e.g., 53)
type DNSRules map[uint16]IPRules

// IPRules is an unsorted collection of IPrules
type IPRules []IPRule

// IPRule stores the allowed destination IPs for a DNS names matching a regex
type IPRule struct {
    Re  RuleRegex
    IPs map[string]struct{} // IPs, nil set is wildcard and allows all IPs!
}

// RuleRegex is a wrapper for *regexp.Regexp so that we can define marshalers for it.
type RuleRegex struct {
    *regexp.Regexp
}
