package main

import (
	_ "embed"
	"time"
)

//go:embed chart.html
var chartTemplateSource string

// Generate graphs for use in the report if required. Return a json dataset that can be used for data size and
// throughput graphing

type DataSet struct {
	Data []Item
	Id   string
	Name string
}

type Item struct {
	TimeStamp time.Time
	Value     int
}

/*

func initialiseGraph(name, id string, filter func(*ConfigEvent) bool, events []*ConfigEvent) *DataSet {
	dataset := DataSet{Id: id, Name: name, Data: []Item{}}
	for _, event := range events {
		if filter(event) {
			dataset.Data = append(dataset.Data, Item{TimeStamp: event.TimeStamp, Value: event.To})
		}
	}

	return &dataset
}

func lastTS(datasets []*DataSet) time.Time {
	last := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, ds := range datasets {
		for _, item := range ds.Data {
			if item.TimeStamp.After(last) {
				last = item.TimeStamp
			}
		}
	}

	return last
}

func firstTS(datasets []*DataSet) time.Time {
	first := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, ds := range datasets {
		for _, item := range ds.Data {
			if item.TimeStamp.Before(first) {
				first = item.TimeStamp
			}
		}
	}

	return first
}

// Make sure we have sanely graphable data for a database by iterating finding the start date for a db
// and looking for a database size change within two minutes of the DB activated event. If not, set to
// 1mb unless database info indicates it's fixed when we'll be clearing it out anyway.
func fixupDBs(datasets []*DataSet, events []*ConfigEvent, dbInfo []*rcutils.DBInfo) []DataSet {

	flexible := []*DataSet{}

	fixed := map[string]string{}
	for _, db := range fixedDbs(dbInfo) {
		fixed[db] = db
	}

	// Build a list of databases with flexible type behaviour (size and throughput values)
	for _, dataset := range datasets {
		if _, ok := fixed[dataset.Name]; !ok {
			flexible = append(flexible, dataset)
		}
	}

	// For each of those we need to find if it has zero or one value in the dataset.
	// If it does we need to find the right value to add and set it as the first and last
	// value if zero and the the last if not.

}

// Given the databases.csv file return a list of databases that are in fixed subs. We are only
// interested in real fixed so the Private endpoint must be empty.
func fixedDbs(dbInfo []*rcutils.DBInfo) []string {
	fixed := []string{}
	for _, db := range dbInfo {
		if db.Fixed() && db.PrivateEndpoint == "N/A" {
			fixed = append(fixed, db.DatabaseName)
		}
	}
	return fixed
}

/*
func memoryGraph(events []*ConfigEvent, dbInfo []*rcutils.DBInfo) DataSet {

	datasets := []DataSet{}

	for _, event := range events {
		if event.Icon == "memory" {
			item := Item{TimeStamp: event.TimeStamp, Value: event.To}
			graphItem := GraphItem{Data: item, Name: event.Database, Type: "time"}}
			if series, ok := dataset[event.Database]; ok {
				dataset[event.Database] = append(event.Database, GraphItem)
			} else {
				dataset[event.Database] = []GraphItem{graphItem}
			}
		}
	}
}

func throughputGraph(events []*ConfigEvent) (string, error) {

}
*/
