package domain

import "strings"

const (
	ObjectCreatedFilter = "s3:ObjectCreated:*"
	ObjectRemovedFilter = "s3:ObjectRemoved:*"
	PrefixFilter        = "prefix"
	SuffixFilter        = "suffix"
)

type FilterRule struct {
	Name  string
	Value string
}

func (f FilterRule) FilterKey(key string) bool {
	if f.Name == PrefixFilter {
		return strings.HasPrefix(key, f.Value)
	}

	if f.Name == SuffixFilter {
		return strings.HasSuffix(key, f.Value)
	}

	panic("expected FilterRule Name to be prefix or suffix but was " + f.Name)
}

type S3Key struct {
	FilterRules []FilterRule `xml:"FilterRule"`
}

type Filter struct {
	S3Key S3Key
}

func (f Filter) FilterEvents(i interface{}) bool {
	event := i.(NotificationEvent)

	for _, f := range f.S3Key.FilterRules {
		if !f.FilterKey(event.Key) {
			return false
		}
	}

	return true
}
