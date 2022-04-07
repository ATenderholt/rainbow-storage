package domain_test

import (
	"encoding/xml"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"testing"
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
