package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goslogan/rcutils"
)

type DBInfo struct {
	Database      string
	Created       time.Time
	Deleted       time.Time
	InitialOps    int
	InitialSizeMB int
	Ambiguous     bool
	Info          *rcutils.DBStatusInfo
	Events        []*ConfigEvent
}

type ConfigEvent struct {
	TimeStamp time.Time `csv:"date"`
	Database  string    `csv:"database name"`
	Change    string    `csv:"description"`
	Icon      string    `csv:"-"`
	Title     string    `csv:"-"`
	Direction string    `csv:"-"`
	From      int       `csv:"-"`
	To        int       `csv:"-"`
}

type EventMatcher struct {
	match     func(*ConfigEvent) bool
	direction func(*ConfigEvent) string
	icon      func(*ConfigEvent) string
	values    func(*ConfigEvent) (int, int)
	Icon      string
	Title     string
	Direction string
}

var subnetRegex = regexp.MustCompile(`(?i)Source ip/subnet (?:deleted|added). Ip/subnet - ([\d./]+)`)
var sizeRegex = regexp.MustCompile(`(?i)Memory Limit changed from ([\d.]+) ([a-zA-Z]+) to ([\d.]+) ([a-zA-Z]+)`)
var throuphputRegex = regexp.MustCompile(`(?i)Database Throughput was changed from (\d+) ops/sec to (\d+) ops/sec`)

var EventMatchers = []EventMatcher{
	{
		match: func(e *ConfigEvent) bool {
			return subnetRegex.MatchString(e.Change)
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(e.Change, "deleted") {
				return "down"
			} else {
				return "up"
			}
		},
		Icon:  "sign-intersection-side",
		Title: "Network Change",
	},
	{ // match memory size change
		match: func(e *ConfigEvent) bool {
			return sizeRegex.MatchString(e.Change)
		},
		values: func(e *ConfigEvent) (int, int) {
			matches := sizeRegex.FindStringSubmatch(e.Change)

			from, err := convertSize(matches[1], matches[2])
			if err != nil {
				log.Printf("unable to extract sizes from %s", e.Change)
				return 0, 0
			}
			to, err := convertSize(matches[3], matches[4])
			if err != nil {
				log.Printf("unable to extract sizes from %s", e.Change)
				return 0, 0
			}

			return from, to
		},
		direction: func(e *ConfigEvent) string {

			if e.To > e.From {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "memory",
		Title: "Memory Limit Change",
	},
	{ // match db activated/deleted
		match: func(e *ConfigEvent) bool {
			return e.Change == "DB activated" || e.Change == "DB deleted"
		},
		direction: func(e *ConfigEvent) string {
			if e.Change == "DB activated" {
				return "up"
			} else {
				return "down"
			}
		},
		icon: func(e *ConfigEvent) string {
			if e.Change == "DB activated" {
				return "database-fill-check"
			} else {
				return "database-fill-slash"
			}
		},
		Icon:  "database-fill",
		Title: "Database Activation/Deletions",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "db name changed")
		},
		Direction: "NA",
		Icon:      "database-fill-check",
		Title:     "Database Change",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "persistence")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "disabled") {
				return "down"
			}
			return "up"
		},
		Icon:  "shield-exclamation",
		Title: "Persistence Change",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(e.Change, "Cluster enabled")
		},
		Icon:      "hdd-rack-fill",
		Direction: "up",
		Title:     "Clustering enabled",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "replication policy")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "to enabled") {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "share-fill",
		Title: "Replication Change",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "sync source")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "added") {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "symmetry-horizontal",
		Title: "Sync Change",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "sync lag is changed") ||
				strings.Contains(strings.ToLower(e.Change), "connections limit is changed") ||
				strings.Contains(strings.ToLower(e.Change), "alert is changed")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "active - true") {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "envelope",
		Title: "Alerts",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "backup")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "enabled") {
				return "up"
			} else if strings.Contains(strings.ToLower(e.Change), "disabled") {
				return "down"
			} else {
				return "NA"
			}
		},
		Icon:  "cloud-download",
		Title: "Backups",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "module")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "loaded") {
				return "up"
			} else {
				return "NA"
			}
		},
		Icon:  "code-square",
		Title: "Modules",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(e.Change, "Cluster rule")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(e.Change, "added") {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "regex",
		Title: "Cluster Rules",
	},
	{
		match: func(e *ConfigEvent) bool {
			return throuphputRegex.MatchString(e.Change)
		},
		values: func(e *ConfigEvent) (int, int) {
			matches := throuphputRegex.FindStringSubmatch(e.Change)
			from, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Printf("Unable to convert ops/sec of %s to integer - %v", matches[1], err)
				return 0, 0
			}
			to, err := strconv.Atoi(matches[2])
			if err != nil {
				log.Printf("Unable to convert ops/sec of %s to integer - %v", matches[2], err)
				return 0, 0
			}
			return from, to
		},
		direction: func(e *ConfigEvent) string {

			if e.To > e.From {
				return "up"
			} else if e.From > e.To {
				return "down"
			} else {
				return "NA"
			}
		},
		Icon:  "speedometer",
		Title: "Throughput Change",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "eviction policy changed")
		},
		Icon:      "door-open",
		Title:     "Eviction",
		Direction: "NA",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "default redis user")
		},
		Icon:      "lock-fill",
		Title:     "Default Password Change",
		Direction: "NA",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "vpc peering")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "delete") {
				return "down"
			} else if strings.Contains(strings.ToLower(e.Change), "initiated") {
				return "up"
			} else {
				return "NA"
			}
		},
		Icon:  "link-45deg",
		Title: "VPC Peering",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "added infrastructure")
		},
		direction: func(e *ConfigEvent) string {
			return "up"
		},
		Icon:  "cloud-plus",
		Title: "Infrastructure",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "api secret key")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "assigned") {
				return "up"
			}
			return "down"
		},
		Icon:  "key",
		Title: "API Key",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "api access")
		},
		direction: func(e *ConfigEvent) string {
			if strings.Contains(strings.ToLower(e.Change), "enabled") {
				return "up"
			}
			return "down"
		},
		Icon:  "key-fill",
		Title: "API Access",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "oss cluster")
		},
		Direction: "up",
		Icon:      "boxes",
		Title:     "OSS Cluster API",
	},
	{
		match: func(*ConfigEvent) bool {
			return true
		},
		direction: func(*ConfigEvent) string {
			return "NA"
		},
		Icon:  "info-circle-fill",
		Title: "Other Change",
	},
}

func DBInfoFromRCutils(d *rcutils.DBConfigInfo) *DBInfo {
	i :=
		DBInfo{
			Database:      d.Database,
			Created:       d.Created,
			Deleted:       d.Deleted,
			InitialOps:    d.InitialOps,
			InitialSizeMB: d.InitialSizeMB,
			Ambiguous:     d.Ambiguous,
			Info:          d.Info,
			Events:        make([]*ConfigEvent, len(d.Events)),
		}

	for n, e := range d.Events {
		i.Events[n] = ConfigEventFromLogEvent(e)
	}

	return &i
}

func ConfigEventFromLogEvent(l *rcutils.LogEvent) *ConfigEvent {
	e := ConfigEvent{
		TimeStamp: l.TimeStamp,
		Database:  l.Database,
		Change:    l.Change,
	}
	return Match(&e)
}

func Match(e *ConfigEvent) *ConfigEvent {
	for _, m := range EventMatchers {
		if m.match(e) {
			if m.values != nil {
				e.From, e.To = m.values(e)
			}
			if m.icon != nil {
				e.Icon = m.icon(e)
			} else {
				e.Icon = m.Icon
			}
			if m.direction != nil {
				e.Direction = m.direction(e)
			} else {
				e.Direction = m.Direction
			}
			e.Title = m.Title
			break
		}
	}

	return e
}

func convertSize(num, scale string) (int, error) {
	base, err := strconv.ParseFloat(num, 32)
	if err != nil {
		return 0, err
	}
	if strings.ToLower(scale) == "gb" {
		return int(base * 1000), nil
	}
	if strings.ToLower(scale) == "mb" {
		return int(base), nil
	}
	return int(base), nil
}
