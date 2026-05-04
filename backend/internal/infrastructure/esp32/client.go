package esp32

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type response struct {
	Status string             `json:"status"`
	Data   domain.MotorStatus `json:"data"`
}

type MotorClient struct {
	config *domain.MotorConfig
	http   *http.Client
}

func NewMotorClient(c *domain.MotorConfig) *MotorClient {
	return &MotorClient{
		config: c,
		http:   &http.Client{},
	}
}

func (c *MotorClient) GetConfig(ctx context.Context) (*domain.MotorConfig, error) {
	url := fmt.Sprintf("http://%s/config", c.config.CurrentIP)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[MOTOR-%d] %s  cannot be reached", c.config.MotorID, url)
	}
	defer resp.Body.Close()

	var res domain.MotorConfig
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("[MOTOR-%d] %s  returns invalid data format", c.config.MotorID, url)
	}
	res.CurrentIP = c.config.CurrentIP

	c.config = &res

	return &res, nil
}

func (c *MotorClient) GetStatus(ctx context.Context) (*domain.MotorStatus, error) {
	url := fmt.Sprintf("http://%s/status", c.config.CurrentIP)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[MOTOR-%d] %s  cannot be reached", c.config.MotorID, url)
	}
	defer resp.Body.Close()

	var res response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("[MOTOR-%d] %s returns invalid data format", c.config.MotorID, url)
	}

	return &res.Data, nil
}

func (c *MotorClient) Move(ctx context.Context, steps int, speed float64) error {
	url := fmt.Sprintf("http://%s/move?steps=%d&speed=%f", c.config.CurrentIP, steps, speed)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("[MOTOR-%d] %s cannot form request", c.config.MotorID, url)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("[MOTOR-%d] %s  cannot be reached", c.config.MotorID, url)
	}
	defer resp.Body.Close()

	return nil
}

func (c *MotorClient) Stop(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/stop", c.config.CurrentIP)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("[MOTOR-%d] %s cannot form request", c.config.MotorID, url)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("[MOTOR-%d] %s  cannot be reached", c.config.MotorID, url)
	}
	defer resp.Body.Close()

	return nil

}
