package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/forfuncsake/powerpal"
)

func main() {
	server := flag.String("s", "http://127.0.0.1:8086", "url of the influxdb server")
	org := flag.String("o", "home", "name of the influxdb org to query and write")
	bucket := flag.String("b", "data", "name of the influxdb bucket to query and write")
	measurement := flag.String("m", "powerpal", "name of the influxdb measurement to query and write")
	flag.Parse()

	serial := os.Getenv("POWERPAL_DEVICE_SERIAL")
	token := os.Getenv("POWERPAL_API_TOKEN")
	dbtoken := os.Getenv("INFLUX_API_TOKEN")

	p := powerpal.NewClient("https://readings.powerpal.net", serial, token)

	c := influxdb2.NewClient(*server, dbtoken)
	defer c.Close()

	q := c.QueryAPI(*org)
	query := fmt.Sprintf(`from(bucket:"%s") |> range(start:-365d) |> filter(fn: (r) => r._measurement == "%s") |> last()`, *bucket, *measurement)
	result, err := q.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	if result.Err() != nil {
		log.Fatalf("query parsing error: %s\n", result.Err().Error())
	}

	//TODO: set to zero after we have batched writes
	//lastRecord := time.Unix(0,0)
	lastRecord := time.Now().Add(-time.Hour)
	if result.Next() {
		lastRecord = result.Record().Time()
	} else {
		log.Printf("No last record found, fetching all data after %s...", lastRecord)
	}

	readings, err := p.FetchReadings(lastRecord.Add(time.Second))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Retrieved %d new readings", len(readings))
	if len(readings) == 0 {
		return
	}

	fmt.Println("writing to influxdb")
	e := powerpal.NewInfluxExporter(c, *org, *bucket, *measurement)
	if err := e.Write(readings); err != nil {
		log.Fatal(err)
	}
}
