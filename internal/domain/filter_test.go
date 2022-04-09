package domain_test

import (
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterEventsNoFilters(t *testing.T) {
	filter := domain.Filter{
		S3Key: domain.S3Key{
			FilterRules: []domain.FilterRule{},
		},
	}

	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.bin"}))
	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.txt"}))
	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test2.bin"}))
}

func TestFilterEventsSuffixOnly(t *testing.T) {
	filter := domain.Filter{
		S3Key: domain.S3Key{
			FilterRules: []domain.FilterRule{
				{
					Name:  domain.SuffixFilter,
					Value: "bin",
				},
			},
		},
	}

	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.bin"}))
	assert.False(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.txt"}))
	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test2.bin"}))
}

func TestFilterEventsPrefixOnly(t *testing.T) {
	filter := domain.Filter{
		S3Key: domain.S3Key{
			FilterRules: []domain.FilterRule{
				{
					Name:  domain.PrefixFilter,
					Value: "test1",
				},
			},
		},
	}

	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.bin"}))
	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.txt"}))
	assert.False(t, filter.FilterEvents(domain.NotificationEvent{Key: "test2.bin"}))
}

func TestFilterEventsPrefixAndSuffix(t *testing.T) {
	filter := domain.Filter{
		S3Key: domain.S3Key{
			FilterRules: []domain.FilterRule{
				{
					Name:  domain.PrefixFilter,
					Value: "test1",
				},
				{
					Name:  domain.SuffixFilter,
					Value: "bin",
				},
			},
		},
	}

	assert.True(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.bin"}))
	assert.False(t, filter.FilterEvents(domain.NotificationEvent{Key: "test1.txt"}))
	assert.False(t, filter.FilterEvents(domain.NotificationEvent{Key: "test2.bin"}))
}
