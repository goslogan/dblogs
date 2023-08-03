package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/dimchansky/utfbom"
	"github.com/gocarina/gocsv"
	"github.com/spf13/pflag"
)

var accountIds []uint
var subscriptionIds []uint
var sources []string
var databases []string
var dbSort bool

type sourceInfo struct {
	reader         io.Reader
	accountId      uint
	subscriptionId uint
	name           string
}

type ConfigEvent struct {
	TimeStamp time.Time `csv:"date"`
	Database  string    `csv:"database name"`
	Change    string    `csv:"description"`
	Activity  string    `csv:"activity"`
}

type ConfigEvents []ConfigEvent

func main() {
	pflag.Parse()
	sources := initSources()
	events := []*ConfigEvent{}

	for _, source := range sources {
		sEvents := []*ConfigEvent{}

		errHandler := func(e *csv.ParseError) bool {
			log.Printf("Parse error in %s - %v", source.name, e)
			return true
		}

		err := gocsv.UnmarshalWithErrorHandler(source.reader, errHandler, &sEvents)
		if err != nil {
			log.Fatalf("Parse error in %s - %v", source.name, err)
		}

		for _, event := range sEvents {
			if event.Activity == "Configuration" {
				ok := true
				if len(databases) > 0 {
					ok = false
					for _, db := range databases {
						if event.Database == db {
							ok = true
							break
						}
					}
				}
				if ok {
					events = append(events, event)
				}
			}
		}
	}

	if len(events) > 1 || dbSort {
		sort.Slice(events, func(i, j int) bool {
			if dbSort && events[i].Database != events[j].Database {
				return events[i].Database < events[j].Database
			} else {
				return events[i].TimeStamp.Before(events[j].TimeStamp)
			}
		})
	}

	for _, event := range events {
		fmt.Printf("%s: %s: %s\n", event.TimeStamp.Format(time.RFC3339), event.Database, event.Change)
	}

}

func init() {
	pflag.BoolVarP(&dbSort, "dbsort", "b", false, "sort by database name before timestamp")
	pflag.StringSliceVarP(&sources, "file", "f", []string{}, "list of csv files to process")
	pflag.UintSliceVarP(&accountIds, "accounts", "a", []uint{0}, "account ids matching sources")
	pflag.UintSliceVarP(&subscriptionIds, "subscriptions", "s", []uint{0}, "subscription ids matching sources")
	pflag.StringSliceVarP(&databases, "databases", "d", []string{}, "report only these named databases")
}

func initSources() []sourceInfo {
	sourceCount := len(sources)
	if sourceCount == 0 {
		sourceCount = 1
	}
	sourceReaders := make([]sourceInfo, sourceCount)

	if len(sources) == 0 {
		sourceReaders[0].reader = utfbom.SkipOnly(os.Stdin)
		sourceReaders[0].accountId = accountIds[0]
		sourceReaders[0].subscriptionId = subscriptionIds[0]
		sourceReaders[0].name = "STDIN"
	} else {
		for n, source := range sources {
			var a, s uint
			if n >= len(accountIds) {
				a = accountIds[n]
			}
			if n >= len(subscriptionIds) {
				s = subscriptionIds[n]
			}
			if reader, err := os.Open(source); err != nil {
				log.Fatalf("Unable to open %s - %v", source, err)
			} else {
				sourceReaders[n].reader = utfbom.SkipOnly(reader)
				sourceReaders[n].accountId = a
				sourceReaders[n].subscriptionId = s
				sourceReaders[n].name = source
			}
		}
	}

	return sourceReaders

}

func (c ConfigEvents) Len() int      { return len(c) }
func (c ConfigEvents) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c ConfigEvents) Less(i, j int) bool {
	if dbSort && c[i].Database != c[j].Database {
		return c[i].Database < c[j].Database
	} else {
		return c[i].TimeStamp.Before(c[j].TimeStamp)
	}
}
