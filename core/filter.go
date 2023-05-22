package core

import "strings"

// A prefix of a type obvious has to be of the same type.
//type PrefixId EventId

type Filter struct {

    // An id here can be a prefix string to an event ID.
    Ids []string `json:"ids,omitempty"`

	// Only broadcast messages with kind in list.
	Kinds []uint32 `json:"kinds,omitempty"`
}

type Filters []Filter

func (s Filters) Match(event *Event) bool {
	for _, f := range s {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

func (s Filter) Matches(event *Event) bool {

	if event == nil {
		return false
	}

	if s.Ids != nil && !containsPrefix(s.Ids, event.Id) {
		return false
	}

	if s.Kinds != nil && !contains(s.Kinds, event.Kind) {
		return false
	}

	return true
}

func containsPrefix(prefixlist []string, id EventId) bool {
	for _, prefix := range prefixlist {
		if strings.HasPrefix(string(id), prefix) {
			return true
		}
	}
	return false
}

func contains(s []uint32, target uint32) bool {
	for _, item := range s {
		if item == target {
			return true
		}
	}
	return false
}
