package config

import (
	"slices"
	"strings"
)

type SourceType string

// Source Location
const (
	SourceWhizzy SourceType = "whizzy"
	SourceDavid  SourceType = "david"
)

// Source Location
var AllSources = []SourceType{
	SourceWhizzy,
	SourceDavid,
}

func (s SourceType) IsValid() bool {
	return slices.Contains(AllSources, s)
}

func AllowedSourcesString() string {
	strSources := make([]string, len(AllSources))
	for i, s := range AllSources {
		strSources[i] = string(s)
	}
	return strings.Join(strSources, ", ")
}
