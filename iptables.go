package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

type IPTablesRule struct {
	table    string
	chain    string
	rulespec []string
}

func rules(n, hostDevice, virtualDevice string) []IPTablesRule {

	return []IPTablesRule{
		// iptables -t nat -A POSTROUTING -s ${DEVICE_ADDR}/24 -o eth0 -j MASQUERADE
		{"nat", "POSTROUTING", []string{"-s", n, "-o", hostDevice, "-j", "MASQUERADE"}},
		// iptables -A FORWARD -i eth0 -o ${DEVICE_NAME} -j ACCEPT
		{"filter", "FORWARD", []string{"-i", hostDevice, "-o", virtualDevice, "-j", "ACCEPT"}},
		// iptables -A FORWARD -o eth0 -i ${DEVICE_NAME} -j ACCEPT
		{"filter", "FORWARD", []string{"-o", hostDevice, "-i", virtualDevice, "-j", "ACCEPT"}},
	}
}

func setIptables() error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("new iptable instance err: %v", err)
	}

	for _, rule := range rules(bridgeIP, hostDevice, bridgeName) {
		log.Println("Adding iptables rule: ", strings.Join(rule.rulespec, " "))
		if err := ipt.AppendUnique(rule.table, rule.chain, rule.rulespec...); err != nil {
			return fmt.Errorf("failed to insert IPTables rule: %v", err)
		}
	}
	return nil
}
