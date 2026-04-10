package main

import (
	"net/netip"
	"os/exec"
	"strings"
)

func getPrimaryIP() (netip.Addr, error) {
	cmd := exec.Command(libexecPath("get-primary-ip"))
	output, err := cmd.Output()
	if err != nil {
		return netip.Addr{}, err
	}
	return netip.ParseAddr(strings.TrimSpace(string(output)))
}
