package main

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
)

type preflightResult struct {
	PublicIP string
	Domain   string
}

func requireRoot() error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("cannot determine current user: %w", err)
	}
	if u.Uid != "0" {
		return fmt.Errorf("root access required — re-run with sudo")
	}
	return nil
}

func runPreflight(email, domain string) (*preflightResult, error) {
	phaseLog(1, "Running preflight checks...")

	if err := requireRoot(); err != nil {
		return nil, err
	}
	stepLog("Running as root")

	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("unsupported OS: %s (only linux is supported)", runtime.GOOS)
	}
	arch := runtime.GOARCH
	if arch != "amd64" && arch != "arm64" {
		return nil, fmt.Errorf("unsupported architecture: %s (only amd64 and arm64)", arch)
	}
	stepLog(fmt.Sprintf("OS: linux/%s", arch))

	checkResources()

	k3sRunning := isK3sRunning()
	for _, port := range []int{80, 443, 6443} {
		if isPortInUse(port) {
			if k3sRunning {
				continue
			}
			return nil, fmt.Errorf("port %d is already in use (not by k3s)", port)
		}
	}
	if k3sRunning {
		stepLog("k3s already running -- port checks skipped for managed ports")
	} else {
		stepLog("Ports 80, 443, 6443 are free")
	}

	if !commandExists("curl") {
		return nil, fmt.Errorf("curl is required but not found")
	}

	publicIP, err := detectPublicIP()
	if err != nil {
		return nil, fmt.Errorf("detecting public IP: %w", err)
	}
	stepLog(fmt.Sprintf("Public IP: %s", publicIP))

	if domain == "" {
		domain = fmt.Sprintf("stackdome.%s.nip.io", publicIP)
		stepLog(fmt.Sprintf("Using default domain: %s", domain))
	} else {
		stepLog(fmt.Sprintf("Using custom domain: %s", domain))
	}

	stepLog(fmt.Sprintf("Admin email: %s", email))
	successLog("Preflight checks passed")
	return &preflightResult{PublicIP: publicIP, Domain: domain}, nil
}

func detectPublicIP() (string, error) {
	ip, err := output("curl", "-4", "-s", "--max-time", "10", "ifconfig.me")
	if err != nil {
		return "", fmt.Errorf("failed to detect public IP (curl ifconfig.me): %w", err)
	}
	ip = strings.TrimSpace(ip)
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid public IP response")
	}
	return ip, nil
}

func isK3sRunning() bool {
	err := run("systemctl", "is-active", "--quiet", "k3s")
	return err == nil
}

func isPortInUse(port int) bool {
	out, err := output("ss", "-tlnp")
	if err != nil {
		return false
	}
	return strings.Contains(out, fmt.Sprintf(":%d ", port))
}

func checkResources() {
	cpuCount := runtime.NumCPU()
	if cpuCount < 2 {
		warnLog(fmt.Sprintf("Only %d CPU(s) detected -- 2+ recommended", cpuCount))
	} else {
		stepLog(fmt.Sprintf("CPUs: %d", cpuCount))
	}

	memBytes, err := readMemTotal()
	if err != nil {
		warnLog("Could not detect system memory")
	} else {
		memGB := memBytes / (1024 * 1024 * 1024)
		if memGB < 4 {
			warnLog(fmt.Sprintf("Only %d GB RAM detected -- 4+ GB recommended", memGB))
		} else {
			stepLog(fmt.Sprintf("Memory: %d GB", memGB))
		}
	}

	diskFreeGB, err := readDiskFree("/")
	if err != nil {
		warnLog("Could not detect available disk space")
	} else if diskFreeGB < 20 {
		warnLog(fmt.Sprintf("Only %d GB disk space available -- 20+ GB recommended", diskFreeGB))
	} else {
		stepLog(fmt.Sprintf("Disk: %d GB available", diskFreeGB))
	}
}

func readDiskFree(path string) (uint64, error) {
	out, err := output("df", "--output=avail", "-B1", path)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}
	avail, err := strconv.ParseUint(strings.TrimSpace(lines[1]), 10, 64)
	if err != nil {
		return 0, err
	}
	return avail / (1024 * 1024 * 1024), nil
}

func readMemTotal() (uint64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return kb * 1024, nil
			}
		}
	}
	return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
}
