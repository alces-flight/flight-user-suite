package main

import (
	"net/netip"
	"os/exec"
	"strings"

	"charm.land/log/v2"
)

func getPrimaryIP() (netip.Addr, error) {
	cmd := exec.Command(libexecPath("get-primary-ip"))
	output, err := cmd.Output()
	if err != nil {
		return netip.Addr{}, err
	}
	return netip.ParseAddr(strings.TrimSpace(string(output)))
}

func portIsReachable(port int) bool {
	cmd := exec.Command(libexecPath("reachable"))
	output, err := cmd.Output()
	if err != nil {
		log.Debug("unable to determine reachability", "port", port, "err", err)
	}
	return strings.TrimSpace(string(output)) == "true"
}
