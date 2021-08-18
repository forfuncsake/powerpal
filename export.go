package powerpal

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Exporter interface {
	Write([]MeterReading) error
}

// Assert that our exporters all satisfy the Exporter interface
var _ Exporter = &CSVExporter{}
var _ Exporter = &InfluxExporter{}

type CSVExporter struct {
	writer      io.Writer
	writeHeader bool
	allFields   bool
}

// NewCSVExporter returns an exporter configured to write csv data. It can be configured
// to write or omit (when appending) a header row; and to emit the same fields as the
// powerpal app CSV export (datetime_utc,watt_hours,cost_dollars,is_peak), or all fields.
func NewCSVExporter(w io.Writer, writeHeader bool, allFields bool) *CSVExporter {
	return &CSVExporter{
		writer:      w,
		writeHeader: writeHeader,
		allFields:   allFields,
	}
}

// Write implements the Exporter interface for CSV output.
func (e *CSVExporter) Write(readings []MeterReading) error {
	toStrings := func(rr []MeterReading) [][]string {
		lines := make([][]string, 0, len(rr)+1)
		if e.writeHeader {
			line := []string{
				"datetime_utc",
				"watt_hours",
				"cost_dollars",
				"is_peak",
			}
			if e.allFields {
				line = append(line,
					"pulses",
					"samples",
				)
			}
			lines = append(lines, line)
		}
		for _, r := range rr {
			line := []string{
				time.Unix(r.Timestamp, 0).UTC().Format("2006-01-02 15:04:05"),
				strconv.Itoa(r.WattHours),
				strconv.FormatFloat(r.Cost, 'f', 10, 64),
				strconv.FormatBool(r.Peak),
			}
			if e.allFields {
				line = append(line,
					strconv.Itoa(r.Pulses),
					strconv.Itoa(r.Samples),
				)
			}
			lines = append(lines, line)
		}

		return lines
	}

	w := csv.NewWriter(e.writer)
	if err := w.WriteAll(toStrings(readings)); err != nil {
		return fmt.Errorf("failed to write csv: %w", err)
	}

	if err := w.Error(); err != nil {
		return fmt.Errorf("csv writer error: %w", err)
	}

	return nil
}

type InfluxExporter struct {
	client      influxdb2.Client
	org         string
	bucket      string
	measurement string
}

// NewInfluxExporter returns an exporter configured to write results to InfluxDB.
func NewInfluxExporter(c influxdb2.Client, org string, bucket string, measurement string) *InfluxExporter {
	return &InfluxExporter{
		client:      c,
		org:         org,
		bucket:      bucket,
		measurement: measurement,
	}
}

// Write implements the Exporter interface for InfluxDB output.
func (e *InfluxExporter) Write(readings []MeterReading) error {
	api := e.client.WriteAPI(e.org, e.bucket)

	firstErr := make(chan error, 1)
	flushed := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		errorsCh := api.Errors()
		for {
			select {
			case err, ok := <-errorsCh:
				if !ok {
					select {
					case firstErr <- nil:
					default:
					}
					return
				}
				if err != nil {
					select {
					case firstErr <- err:
					default:
					}
					cancel()
				}
			case <-flushed:
				select {
				case firstErr <- nil:
				default:
				}
				return
			}
		}
	}()

LOOP:
	for _, reading := range readings {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}
		p := influxdb2.NewPoint(e.measurement,
			map[string]string{"peak": fmt.Sprint(reading.Peak)},
			map[string]interface{}{
				"cost":       reading.Cost,
				"pulses":     reading.Pulses,
				"samples":    reading.Samples,
				"watt_hours": reading.WattHours,
			},
			time.Unix(reading.Timestamp, 0))
		api.WritePoint(p)
	}
	api.Flush()
	close(flushed)

	return <-firstErr
}
