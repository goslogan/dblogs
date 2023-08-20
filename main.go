package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sort"
	"text/template"
	"time"

	"github.com/dimchansky/utfbom"

	_ "embed"

	"github.com/gocarina/gocsv"
	"github.com/spf13/pflag"
)

//go:embed report.template
var defaultTemplate string

var (
	accountIds, subscriptionIds []uint
	sources, databases          []string
	dbSort, timeline, hourly    bool
	output                      string
	title                       string
	firstDate, lastDate         string
	templatePath                string
)

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
	Icon      string    `csv:"-"`
	Title     string    `csv:"-"`
	Direction string    `csv:"-"`
}

type ConfigLoadEvent struct {
	ConfigEvent
	Activity string `csv:"activity"`
}

const ROWLEN = 5

type ConfigEvents []ConfigEvent

func main() {
	pflag.Parse()
	sources := initSources()
	events := []*ConfigEvent{}

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

	for _, source := range sources {
		sEvents := []*ConfigLoadEvent{}

		errHandler := func(e *csv.ParseError) bool {
			log.Printf("Parse error in %s - %v", source.name, e)
			return true
		}

		err := gocsv.UnmarshalWithErrorHandler(source.reader, errHandler, &sEvents)
		if err != nil {
			log.Fatalf("Parse error in %s - %v", source.name, err)
		}

		for _, event := range sEvents {
			if event.Activity == "Configuration" && event.TimeStamp.After(startDate) && event.TimeStamp.Before(endDate) {
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
					parsedEvent := Match(&ConfigEvent{
						Database:  event.Database,
						Change:    event.Change,
						TimeStamp: event.TimeStamp,
					})
					events = append(events, parsedEvent)
				}
			}
		}
	}

	if len(events) > 1 || dbSort || timeline {
		sort.Slice(events, func(i, j int) bool {
			if dbSort && events[i].Database != events[j].Database {
				return events[i].Database < events[j].Database
			} else {
				return events[i].TimeStamp.Before(events[j].TimeStamp)
			}
		})
	}

	if timeline {
		renderTimeline(events)
	} else {
		dumpEvents(events)
	}
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

func dumpEvents(events []*ConfigEvent) {
	var writer io.Writer
	var err error
	if output == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(output)
		if err != nil {
			log.Fatalf("Unable to write to output file %s - %v", output, err)
		}
	}

	err = gocsv.Marshal(events, writer)
	if err != nil {
		log.Fatalf("Unable to write to output file %s - %v", output, err)
	}
}

func renderTimeline(events []*ConfigEvent) {
	var writer io.Writer
	var err error
	var timeFormat = "2006-01-02"

	if hourly {
		timeFormat = "2006-01-02 15:04"
	}

	timeline := make(map[string]map[string][]*ConfigEvent)
	databases := make(map[string]string)

	for _, event := range events {
		t := event.TimeStamp.Format(timeFormat)
		databases[event.Database] = event.Database
		if event.Database == "" {
			databases[event.Database] = "Subscription"
		}
		if _, ok := timeline[t]; !ok {
			timeline[t] = make(map[string][]*ConfigEvent)
		}
		timeline[t][event.Database] = append(timeline[t][event.Database], event)
	}

	tmpl, err := initTemplate()

	if err != nil {
		log.Fatalf("unable to parse output template - %v", err)
	}

	if output == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(output)
		if err != nil {
			log.Fatalf("Unable to write to output file %s - %v", output, err)
		}
	}

	data := map[string]interface{}{
		"Timeline":  timeline,
		"Title":     title,
		"Databases": databases,
		"Legend":    legend(),
	}
	err = tmpl.Execute(writer, data)

	if err != nil {
		log.Fatalf("unable to execute output template - %v", err)
	}
}

// Get the events for a db at a time.
func Events(databases map[string][]*ConfigEvent, database string) []*ConfigEvent {
	if events, ok := databases[database]; !ok {
		return []*ConfigEvent{}
	} else {
		sort.Slice(events, func(i, j int) bool {
			return events[i].TimeStamp.Before(events[j].TimeStamp)
		})
		return events
	}
}

// Legend - return a function which returns the legend in order and by rows.
func legend() [][]map[string]string {
	rows := len(EventMatchers) / ROWLEN
	if len(EventMatchers)%ROWLEN != 0 {
		rows++
	}

	legend := make([][]map[string]string, rows)

	for n := 0; n < rows; n++ {
		legend[n] = make([]map[string]string, ROWLEN)
		for m := 0; m < ROWLEN; m++ {
			p := (n * ROWLEN) + m
			if p >= len(EventMatchers) {
				break
			}
			legend[n][m] = map[string]string{"Title": EventMatchers[p].Title, "Icon": EventMatchers[p].Icon}
		}
	}

	return legend
}

func initTemplate() (*template.Template, error) {

	var templateSource = defaultTemplate

	if templatePath != "" {
		if content, err := os.ReadFile(templatePath); err != nil {
			log.Fatalf("unable to read template from '%s' - %v", templatePath, err)
		} else {
			templateSource = string(content)
		}
	}
	return template.New("render").Funcs(map[string]any{"Events": Events}).Parse(templateSource)
}

func init() {
	pflag.BoolVarP(&dbSort, "dbsort", "b", false, "sort by database name before timestamp")
	pflag.StringSliceVarP(&sources, "files", "f", []string{}, "list of csv files to process")
	pflag.UintSliceVarP(&accountIds, "accounts", "a", []uint{0}, "account ids matching sources")
	pflag.UintSliceVarP(&subscriptionIds, "subscriptions", "s", []uint{0}, "subscription ids matching sources")
	pflag.StringSliceVarP(&databases, "databases", "d", []string{}, "report only these named databases")
	pflag.BoolVarP(&timeline, "timeline", "t", false, "generate a timeline graph for each database")
	pflag.StringVarP(&output, "output", "o", "", "output file for CSV dump or HTML timeline")
	pflag.BoolVarP(&hourly, "hourly", "h", false, "aggregate hourly instead of daily")
	pflag.StringVarP(&title, "title", "i", "Configuration Timeline", "the title for the timeline report")
	pflag.StringVarP(&firstDate, "from", "F", "", "First date to include in the output (yyyy-mm-dd)")
	pflag.StringVarP(&lastDate, "to", "T", "", "Last date to include in the output (yyyy-mm-dd)")
	pflag.StringVarP(&templatePath, "template", "p", "", "Path to a custom template for output")
}
