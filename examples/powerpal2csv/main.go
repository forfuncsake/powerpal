package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/forfuncsake/powerpal"
)

func main() {
	since := flag.Duration("t", time.Hour, "`duration` of time to look back")
	flag.Parse()

	serial := os.Getenv("POWERPAL_DEVICE_SERIAL")
	token := os.Getenv("POWERPAL_API_TOKEN")

	p := powerpal.NewClient("https://readings.powerpal.net", serial, token)
	readings, err := p.FetchReadings(time.Now().UTC().Add(-1 * *since))
	if err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("powerpal_export_%d.csv", time.Now().Unix())
	log.Printf("Retrieved %d new readings, saving to %s", len(readings), filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	e := powerpal.NewCSVExporter(f, true, true)
	if err := e.Write(readings); err != nil {
		log.Fatal(err)
	}
}
