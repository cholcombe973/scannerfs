package main

import (
	"os/exec"
	"regexp"
)

type Host struct {
	address  string
	nmapData []byte
}

func (host *Host) nmapScan() error {
	scan, err := exec.Command("nmap", host.address).Output()
	host.nmapData = scan
	return err
}

type Scanner struct {
	Device string // Network device used for all scans
	Hosts  map[string]Host
}

func NewScanner(device string) (Scanner, error) {
	return Scanner{Device: device, Hosts: make(map[string]Host)}, nil
}

func (scanner *Scanner) getBroadcastAddress() (string, error) {
	ifconfig, err := exec.Command("ifconfig", scanner.Device).Output()
	if err != nil {
		return "", err
	}
	addr := string(regexp.MustCompile(
		"cast[ :]+([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)").FindSubmatch(ifconfig)[1])
	return addr, nil

}

func (scanner *Scanner) EnumerateHosts() error {
	bcastAddr, err := scanner.getBroadcastAddress()
	ping, err := exec.Command("ping", "-m", "1", "-c", "2", bcastAddr).Output()
	if err != nil {
		return err
	}
	addresses := regexp.MustCompile(
		"[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+").FindAllString(string(ping), -1)
	for _, address := range addresses {
		scanner.Hosts[address] = Host{address: address}
	}
	return nil
}
