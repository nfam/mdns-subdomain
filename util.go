package main

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
)

func hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return hostname, err
	}
	if !strings.HasSuffix(hostname, ".local") {
		hostname += ".local"
	}
	return hostname, nil
}

func hostip() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("no network ip was found")
}

func ipOfInterface(name string) (string, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if ipnet.IP.IsLoopback() && (iface.Flags&net.FlagLoopback) == 0 {
			continue
		}
		if ipnet.IP.To4() == nil {
			continue
		}
		return ipnet.IP.String(), nil
	}
	return "", errors.New("inteface " + name + " has no ip address")
}

func ipv4(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}

	}
	return true
}
