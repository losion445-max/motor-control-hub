package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type ARPScanner struct {
	interfaceName string
	client        *http.Client
}

func NewARPScanner(iface string, timeout time.Duration) *ARPScanner {
	return &ARPScanner{
		interfaceName: iface,
		client:        &http.Client{Timeout: timeout},
	}
}

func (s *ARPScanner) Discover(ctx context.Context) ([]domain.MotorConfig, error) {
	log.Println("[SCANNER] Reading ARP table...")
	ips, err := s.getIpsFromArpTable()
	if err != nil {
		return nil, err
	}
	log.Printf("[SCANNER] Found %d potential candidates in ARP table", len(ips))

	var found []domain.MotorConfig
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go func(targetIP string) {
			defer wg.Done()
			cfg, err := s.fetchConfig(ctx, targetIP)
			if err == nil {
				mu.Lock()
				cfg.CurrentIP = targetIP
				found = append(found, *cfg)
				log.Printf("[SCANNER] [MOTOR-%d] Success! Found at %s", cfg.MotorID, targetIP)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	log.Printf("[SCANNER] Discovery finished. Total motors found: %d", len(found))
	return found, nil
}

func (s *ARPScanner) getIpsFromArpTable() ([]string, error) {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, fmt.Errorf("[SCANNER] Couldn't read arp table: %w", err)
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 6 && fields[5] == s.interfaceName {
			ips = append(ips, fields[0])
		}
	}
	return ips, nil
}

func (s *ARPScanner) fetchConfig(ctx context.Context, ip string) (*domain.MotorConfig, error) {
	url := fmt.Sprintf("http://%s/config", ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp struct {
		Status string             `json:"status"`
		Data   domain.MotorConfig `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Status != "ok" {
		return nil, fmt.Errorf("status not ok: %s", apiResp.Status)
	}

	return &apiResp.Data, nil
}
