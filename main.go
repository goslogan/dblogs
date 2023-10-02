package main

import (
	"io"
	"log"
	"os"
	"time"

	_ "embed"

	"github.com/dimchansky/utfbom"
	"github.com/goslogan/rcutils"
	"github.com/spf13/pflag"
)

//go:embed report.html
var defaultTemplate string

var (
	databases                                                        []string
	graphs, hourly                                                   bool
	source, output, title, firstDate, lastDate, templatePath, dbfile string
	dbInfo                                                           []*rcutils.DBConfigInfo
)

const ROWLEN = 5

type ConfigEvents []ConfigEvent

func main() {

	var logInput, dbInput io.Reader

	pflag.Parse()

	endTime, _ := time.ParseDuration("23h59m59s")

	startDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)

	var err error
	if firstDate != "" {
		if startDate, err = time.Parse(time.DateOnly, firstDate); err != nil {
			log.Fatalf("unable to parse '%s' as an ISO date", firstDate)
		}
	}

	if lastDate != "" {
		if endDate, err = time.Parse(time.DateOnly, lastDate); err != nil {
			log.Fatalf("unable to parse '%s' as an ISO date", lastDate)
		}

		endDate = endDate.Add(endTime)
	}

	var filter = func(e *rcutils.LogEvent) bool {
		if e.Activity == "Configuration" && e.TimeStamp.After(startDate) && e.TimeStamp.Before(endDate) {
			ok := true
			if len(databases) > 0 {
				ok = false
				for _, db := range databases {
					if e.Database == db {
						ok = true
						break
					}
				}
			}
			return ok
		}
		return false
	}

	if source == "" {
		logInput = os.Stdin
		source = "STDIN"
	} else {
		var err error
		logInput, err = os.Open(source)
		if err != nil {
			log.Fatalf("unable to open system log (%s) - %v", source, err)
		}
	}

	if dbfile == "" {
		log.Fatal("unable to continue - no database csv input file provided")
	}
	dbInput, err = os.Open(dbfile)
	if err != nil {
		log.Fatalf("unable to open database csv file (%s) - %v", dbfile, err)
	}

	result, err := rcutils.AccountDatabaseInfo(utfbom.SkipOnly(logInput), utfbom.SkipOnly(dbInput), filter)
	if err != nil {
		log.Fatalf("unable to process input files - %v", err)
	}

	dbInfo := make(map[string]*DBInfo, len(result))
	for n, i := range result {
		dbInfo[n] = DBInfoFromRCutils(i)
	}
	renderTimeline(dbInfo)
}

func init() {
	pflag.StringVarP(&source, "source", "f", "", "system log file to process, defaults to stdin")
	pflag.StringSliceVarP(&databases, "databases", "d", []string{}, "report only these named databases")
	pflag.BoolVarP(&hourly, "hourly", "h", false, "aggregate hourly instead of daily")
	pflag.StringVarP(&title, "title", "i", "Configuration Timeline", "the title for the report")
	pflag.StringVarP(&firstDate, "from", "F", "", "First date to include in the output (yyyy-mm-dd)")
	pflag.StringVarP(&lastDate, "to", "T", "", "Last date to include in the output (yyyy-mm-dd)")
	pflag.StringVarP(&output, "output", "o", "", "output file, defaults to stdout")
	pflag.StringVarP(&templatePath, "template", "p", "", "Path to a custom template for output")
	pflag.BoolVarP(&graphs, "graphs", "g", false, "Include graphs of changes to throughput & size")
	pflag.StringVarP(&dbfile, "dbfile", "b", "", "Databases export file")
}
