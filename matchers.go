package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

type EventMatcher struct {
	match     func(*ConfigEvent) bool
	direction func(*ConfigEvent) string
	icon      func(*ConfigEvent) string
	Icon      string
	Title     string
	Direction string
}

var subnetRegex = regexp.MustCompile(`(?i)Source ip/subnet (?:deleted|added). Ip/subnet - ([\d./]+)`)
var sizeRegex = regexp.MustCompile(`(?i)Memory Limit changed from (\d+) ([a-zA-Z]+) to (\d+) ([a-zA-Z]+)`)
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
		Title: "Network",
	},
	{ // match memory size change
		match: func(e *ConfigEvent) bool {
			return sizeRegex.MatchString(e.Change)
		},
		direction: func(e *ConfigEvent) string {
			matches := sizeRegex.FindStringSubmatch(e.Change)
			from, err := convertSize(matches[1], matches[2])
			if err != nil {
				log.Printf("unable to extract sizes from %s", e.Change)
				return "NA"
			}
			to, err := convertSize(matches[3], matches[4])
			if err != nil {
				log.Printf("unable to extract sizes from %s", e.Change)
				return "NA"
			}
			if to > from {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "memory",
		Title: "Memory Limit",
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
		Title: "Database",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.HasPrefix(strings.ToLower(e.Change), "db name changed")
		},
		Direction: "NA",
		Icon:      "database-fill-check",
		Title:     "Database",
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
		Title: "Persistence",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(e.Change, "Cluster enabled")
		},
		Icon:      "hdd-rack-fill",
		Direction: "up",
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
		Title: "Replication",
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
		Title: "Sync",
	},
	{
		match: func(e *ConfigEvent) bool {
			return strings.Contains(strings.ToLower(e.Change), "alert is changed")
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
			return strings.Contains(strings.ToLower(e.Change), "sync lag is changed")
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
		Direction: "up",
		Icon:      "code-square",
		Title:     "Modules",
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
		Title: "Cluster Rule",
	},
	{
		match: func(e *ConfigEvent) bool {
			return throuphputRegex.MatchString(e.Change)
		},
		direction: func(e *ConfigEvent) string {
			matches := throuphputRegex.FindStringSubmatch(e.Change)
			from, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Printf("Unable to convert ops/sec of %s to integer - %v", matches[1], err)
				return "NA"
			}
			to, err := strconv.Atoi(matches[2])
			if err != nil {
				log.Printf("Unable to convert ops/sec of %s to integer - %v", matches[2], err)
				return "NA"
			}
			if to > from {
				return "up"
			} else {
				return "down"
			}
		},
		Icon:  "speedometer",
		Title: "Throughput",
	},
	{
		match: func(*ConfigEvent) bool {
			return true
		},
		direction: func(*ConfigEvent) string {
			return "NA"
		},
		Icon:  "info-circle-fill",
		Title: "General",
	},
}

func Match(e *ConfigEvent) *ConfigEvent {
	for _, m := range EventMatchers {
		if m.match(e) {
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

func convertSize(num, scale string) (uint, error) {
	base, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	if strings.ToLower(scale) == "gb" {
		return uint(base) * 1000, nil
	}
	if strings.ToLower(scale) == "mb" {
		return uint(base), nil
	}
	return uint(base), nil
}
