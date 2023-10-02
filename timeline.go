package main

import (
	"io"
	"log"
	"os"
	"slices"
	"sort"
	"text/template"
)

func renderTimeline(events map[string]*DBInfo) {
	var writer io.Writer
	var err error
	var timeFormat = "2006-01-02"
	var timeline map[string]map[string][]*ConfigEvent

	if hourly {
		timeFormat = "2006-01-02 15:04"
	}

	timeline = buildTimeline(events, timeFormat)

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
		"Databases": initDBList(events),
		"Title":     title,
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
	if tmpl, err := template.New("render").Funcs(map[string]any{"Events": Events}).Parse(templateSource); err != nil {
		return nil, err
	} else {
		_, err := tmpl.New("graphs").Parse(chartTemplateSource)
		return tmpl, err
	}

}

// Take the events and convert to a map of maps. The top level is the timeline
// itself. The next level is the databases and the level below is an array of
// events for that database at the time.
func buildTimeline(dbInfo map[string]*DBInfo, timeFormat string) map[string]map[string][]*ConfigEvent {

	var timeline map[string]map[string][]*ConfigEvent = make(map[string]map[string][]*ConfigEvent)

	for db, info := range dbInfo {
		for _, event := range info.Events {
			bucket := event.TimeStamp.Format(timeFormat)
			if _, ok := timeline[bucket]; !ok {
				timeline[bucket] = make(map[string][]*ConfigEvent)
			}
			if _, ok := timeline[bucket][db]; !ok {
				timeline[bucket][db] = make([]*ConfigEvent, 0)
			}
			timeline[bucket][db] = append(timeline[bucket][db], event)
		}
	}

	return timeline
}

// Get the list of databases we have present (it's just the keys of map)
func initDBList(dbInfo map[string]*DBInfo) []string {
	keys := []string{}
	for k := range dbInfo {
		keys = append(keys, k)
	}

	slices.Sort(keys)
	return keys
}
