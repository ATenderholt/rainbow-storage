package domain_test

import (
	"context"
	"encoding/xml"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const notificationExample = `<NotificationConfiguration
    xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
    <CloudFunctionConfiguration>
        <Event>s3:ObjectRemoved:*</Event>
        <Event>s3:ObjectCreated:*</Event>
        <Filter>
            <S3Key>
                <FilterRule>
                    <Name>prefix</Name>
                    <Value>AWSLogs/</Value>
                </FilterRule>
                <FilterRule>
                    <Name>suffix</Name>
                    <Value>.log</Value>
                </FilterRule>
            </S3Key>
        </Filter>
        <Id>tf-s3-lambda-20220407133353589300000001</Id>
        <CloudFunction>arn:aws:lambda:us-west-2:271828182845:function:myaws-copy-file</CloudFunction>
    </CloudFunctionConfiguration>
</NotificationConfiguration>`

type Collector struct {
	keys map[string][]string
}

func (c *Collector) Invoke(source string) func(interface{}) {
	if c.keys == nil {
		c.keys = make(map[string][]string)
	}

	items, ok := c.keys[source]
	if !ok {
		items = []string{}
	}

	return func(i interface{}) {
		items = append(items, i.(domain.NotificationEvent).Key)
		c.keys[source] = items
	}
}

func TestNotificationUnmarshall(t *testing.T) {
	var notification domain.NotificationConfiguration
	err := xml.Unmarshal([]byte(notificationExample), &notification)

	if err != nil {
		t.Fatalf("Unable to unmarshall: %v", err)
	}

	if len(notification.CloudFunctionConfigurations) != 1 {
		t.Fatalf("Expected 1 CloudFunctionConfigruations, but got %d", len(notification.CloudFunctionConfigurations))
	}

	events := notification.CloudFunctionConfigurations[0].Events
	if len(events) != 2 {
		t.Fatalf("Expected 2 Events, but got %d", len(events))
	}

	if events[0] != "s3:ObjectRemoved:*" {
		t.Errorf("Expected event %s but got %s", "s3:ObjectRemoved:*", events[0])
	}

	if events[1] != "s3:ObjectCreated:*" {
		t.Errorf("Expected event %s but got %s", "s3:ObjectCreated:*", events[1])
	}
}

func TestSingleNotificationCloudFunctionConfigurationsOnlyCreates(t *testing.T) {
	cfg := domain.NotificationConfiguration{
		CloudFunctionConfigurations: []domain.CloudFunctionConfiguration{
			{
				Events:        []string{domain.ObjectCreatedFilter},
				CloudFunction: "create",
			},
		},
	}

	var c Collector
	ch, ctx := cfg.Start(&c)
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file1.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file2.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectRemovedEvent, Key: "file3.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file4.bin"}}
	close(ch)

	timeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	<-timeout.Done()

	assert.Equal(t, []string{"file1.bin", "file2.bin", "file4.bin"}, c.keys["create"])
}

func TestSingleNotificationCloudFunctionConfigurationsCreatesWithFilters(t *testing.T) {
	cfg := domain.NotificationConfiguration{
		CloudFunctionConfigurations: []domain.CloudFunctionConfiguration{
			{
				Events:        []string{domain.ObjectCreatedFilter},
				CloudFunction: "create",
				Filter: domain.Filter{
					S3Key: domain.S3Key{
						FilterRules: []domain.FilterRule{
							{
								Name:  domain.PrefixFilter,
								Value: "file1",
							},
							{
								Name:  domain.SuffixFilter,
								Value: "bin",
							},
						},
					},
				},
			},
		},
	}

	var c Collector
	ch, ctx := cfg.Start(&c)
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file1.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file1.txt"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file2.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectRemovedEvent, Key: "file3.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file4.bin"}}
	close(ch)

	timeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	<-timeout.Done()

	assert.Equal(t, []string{"file1.bin"}, c.keys["create"])
}

func TestTwoNotificationCloudFunctionConfigurations(t *testing.T) {
	cfg := domain.NotificationConfiguration{
		CloudFunctionConfigurations: []domain.CloudFunctionConfiguration{
			{
				Events:        []string{domain.ObjectCreatedFilter},
				CloudFunction: "create",
			},
			{
				Events:        []string{domain.ObjectRemovedFilter},
				CloudFunction: "delete",
			},
		},
	}

	var c Collector
	ch, ctx := cfg.Start(&c)
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file1.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectCreatedEvent, Key: "file2.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectRemovedEvent, Key: "file3.bin"}}
	ch <- rxgo.Item{V: domain.NotificationEvent{Event: domain.ObjectRemovedEvent, Key: "file4.bin"}}
	close(ch)

	timeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	<-timeout.Done()

	assert.Equal(t, []string{"file1.bin", "file2.bin"}, c.keys["create"])
	assert.Equal(t, []string{"file3.bin", "file4.bin"}, c.keys["delete"])
}
