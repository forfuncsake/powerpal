package powerpal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

const (
	// Powerpal docs note that the meter_readings API is
	// limited to batchd of 50k records at a time.
	BatchSize = 50000

	// Time series data is stored/returned in 60s intervals.
	SecondsPerReading = 60
)

// Device represents the data returned from the readings API device endpoint.
// It is used to determine timestamps for meter_reading data availability.
type Device struct {
	AvailableDays  int     `json:"available_days"`
	FirstTimestamp int64   `json:"first_reading_timestamp"`
	LastCost       float64 `json:"last_reading_cost"`
	LastTimestamp  int64   `json:"last_reading_timestamp"`
	LastWattHours  int     `json:"last_reading_watt_hours"`
	Serial         string  `json:"serial_number"`
	TotalCost      float64 `json:"total_cost"`
	TotalReads     int     `json:"total_meter_reading_count"`
	TotalWattHours int     `json:"total_watt_hours"`
}

// A MeterReading provides a single interval meter reading snapshot.
type MeterReading struct {
	Cost      float64 `json:"cost"`
	Peak      bool    `json:"is_peak"`
	Pulses    int     `json:"pulses"`
	Samples   int     `json:"samples"`
	Timestamp int64   `json:"timestamp"`
	WattHours int     `json:"watt_hours"`
}

// A Client provides methods to query the Powerpal API server.
type Client struct {
	http *http.Client

	server string
	serial string
	token  string
}

// NewClient creates a Powerpal API client.
func NewClient(addr string, serial string, token string) *Client {
	return &Client{
		http: &http.Client{
			Timeout:   5 * time.Second,
			Transport: &http2.Transport{},
		},

		server: addr,
		serial: serial,
		token:  token,
	}
}

// FetchReadings will fetch all meter readings from the Powerpal API
// since (and including) the provided timestamp.
func (c *Client) FetchReadings(since time.Time) ([]MeterReading, error) {
	// Query device endpoint to get all timestamps
	var device Device
	if err := c.apiCall(c.deviceURL(), &device); err != nil {
		return nil, err
	}

	start := since.Unix()
	if device.LastTimestamp == 0 || device.LastTimestamp <= start {
		return nil, nil
	}

	if device.FirstTimestamp > start {
		start = device.FirstTimestamp
	}
	// Add 1 because we ingest both the first and last timestamps
	count := ((device.LastTimestamp - start) / SecondsPerReading) + 1

	// In case we were fed a start time that isn't on a 60s reading marker
	// (e.g. last reading +1 sec to avoid a double-read).
	start = device.LastTimestamp - ((count - 1) * SecondsPerReading)

	readings := make([]MeterReading, 0, count)
	for count > 0 {
		end := device.LastTimestamp
		if count > BatchSize {
			end = start + (BatchSize * SecondsPerReading)
		}
		count -= BatchSize

		var batch []MeterReading
		if err := c.apiCall(c.readingsURL(start, end), &batch); err != nil {
			return nil, err
		}

		readings = append(readings, batch...)
		start = end + SecondsPerReading
	}

	return readings, nil
}

func (c *Client) deviceURL() string {
	return c.server + "/api/v1/device/" + c.serial
}

func (c *Client) readingsURL(start, end int64) string {
	return fmt.Sprintf("%s/api/v1/meter_reading/%s?start=%d&end=%d", c.server, c.serial, start, end)
}

// Call the Powerpal API at the provided endpoint and unmarshal the
// returned body into v.
func (c *Client) apiCall(endpoint string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-au")
	req.Header.Set("Authorization", c.token)
	req.Header.Set("User-Agent", "Powerpal/1895 CFNetwork/1240.0.4 Darwin/20.5.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
