package domain

import (
	"context"
	"github.com/reactivex/rxgo/v2"
	"strings"
)

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

type CloudFunction string

func (c CloudFunction) Invoke(f EventFunction) func(interface{}) {
	return func(i interface{}) {
		f(string(c), i)
	}
}

type CloudFunctionConfiguration struct {
	Events        []string `xml:"Event"`
	Filter        Filter
	ID            string
	CloudFunction CloudFunction
}

func (c CloudFunctionConfiguration) CreateObservable(source rxgo.Observable) rxgo.Observable {
	return source.
		Filter(c.FilterEvents).
		Filter(c.Filter.FilterEvents)
}

func (c CloudFunctionConfiguration) FilterEvents(i interface{}) bool {
	event := i.(NotificationEvent)

	for _, filter := range c.Events {
		if strings.HasPrefix(filter, event.Event) {
			return true
		}
	}

	return false
}

type NotificationConfiguration struct {
	CloudFunctionConfigurations []CloudFunctionConfiguration `xml:"CloudFunctionConfiguration"`
}

type EventFunction func(string, interface{})

func (n NotificationConfiguration) Start(f EventFunction) (chan rxgo.Item, context.Context) {
	ch := make(chan rxgo.Item)

	source := rxgo.FromChannel(ch, rxgo.WithPublishStrategy())
	for _, funcConfigs := range n.CloudFunctionConfigurations {
		obs := funcConfigs.CreateObservable(source)
		obs.DoOnNext(funcConfigs.CloudFunction.Invoke(f))
	}

	ctx, _ := source.Connect(context.Background())
	return ch, ctx
}
