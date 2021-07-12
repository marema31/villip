package main

import (
	"sort"

	"github.com/marema31/villip/filter"
	"github.com/marema31/villip/server"
	"github.com/sirupsen/logrus"
)

type filtersList map[string]map[uint8][]*filter.Filter

func insertInFilters(filters filtersList, port string, priority uint8, f *filter.Filter) {
	if _, ok := filters[port]; !ok {
		filters[port] = make(map[uint8][]*filter.Filter)
	}

	if _, ok := filters[port][priority]; !ok {
		filters[port][priority] = make([]*filter.Filter, 0, 1)
	}

	if !f.IsConditional() {
		filters[port][priority] = append(filters[port][priority], f)
	} else {
		// Prepending filter to the list using golang tricks
		filters[port][priority] = append(filters[port][priority], &filter.Filter{})
		copy(filters[port][priority][1:], filters[port][priority])
		filters[port][priority][0] = f
	}
}

func createServer(filters map[uint8][]*filter.Filter, port string, upLog logrus.FieldLogger) *server.Server {
	var s *server.Server

	priorities := make([]uint8, 0, 1)
	for priority := range filters {
		priorities = append(priorities, priority)
	}

	//Sort the priorities in descent order
	sort.Slice(priorities, func(i, j int) bool { return priorities[i] > priorities[j] })

	for _, priority := range priorities {
		for _, f := range filters[priority] {
			if s != nil {
				s.Insert(f)
			} else {
				s = server.New(upLog, port, f)
			}
		}
	}

	return s
}

func createServers(filters filtersList, upLog logrus.FieldLogger) map[string]*server.Server {
	servers := make(map[string]*server.Server)
	for port := range filters {
		servers[port] = createServer(filters[port], port, upLog)
	}

	return servers
}
