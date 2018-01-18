package main

import (
	"fmt"

	"github.com/coreos/go-iptables/iptables"
)

type IPTablesRule struct {
	table    string
	chain    string
	rulespec []string
}

func getIptablesRules(n, hostDevice, virtualDevice string) []IPTablesRule {

	return []IPTablesRule{
		// iptables -t nat -A POSTROUTING -s ${DEVICE_ADDR}/24 -o eth0 -j MASQUERADE
		{"nat", "POSTROUTING", []string{"-s", n, "-o", hostDevice, "-j", "MASQUERADE"}},
		// iptables -A FORWARD -i eth0 -o ${DEVICE_NAME} -j ACCEPT
		{"filter", "FORWARD", []string{"-i", hostDevice, "-o", virtualDevice, "-j", "ACCEPT"}},
		// iptables -A FORWARD -o eth0 -i ${DEVICE_NAME} -j ACCEPT
		{"filter", "FORWARD", []string{"-o", hostDevice, "-i", virtualDevice, "-j", "ACCEPT"}},
	}
}

func teardownIPTables(ipt iptables.IPTables, rules []IPTablesRule) {
	for _, rule := range rules {
		// We ignore errors here because if there's an error it's almost certainly because the rule
		// doesn't exist, which is fine (we don't need to delete rules that don't exist)
		ipt.Delete(rule.table, rule.chain, rule.rulespec...)
	}
}

func setIptables(iptablesRules []IPTablesRule) error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("new iptable instance err: %v", err)
	}

	for _, rule := range iptablesRules {
		if err := ipt.AppendUnique(rule.table, rule.chain, rule.rulespec...); err != nil {
			return fmt.Errorf("failed to insert IPTables rule: %v", err)
		}
	}
	return nil
}
