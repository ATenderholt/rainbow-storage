package domain

import (
	"context"
	"github.com/reactivex/rxgo/v2"
	"strings"
)

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
