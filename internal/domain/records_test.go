package domain_test

import (
	"encoding/json"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const expected = `{
	"eventVersion": "2.1",
	"eventSource": "aws:s3",
	"awsRegion": "us-west-2",
	"eventTime": "2022-04-14T11:39:29.346Z",
	"eventName": "ObjectCreated:CompleteMultipartUpload",
	"userIdentity": {
		"principalId": "AWS:SOMEPRINCIPAL"
	},
	"requestParameters": {
		"sourceIPAddress": "123.45.67.89"
	},
	"responseElements": {
		"x-amz-request-id": "XT6FD2FBQWXM1ABC",
		"x-amz-id-2": "ab7rhq6747Kpa/aBY60gVUd1kd79J7asNC3RvyN6d77zjzYn+aBnTh5107THtwu/qufcgLisDK+30aErdEbk7Rw7a5EokaBC"
	},
	"s3": {
		"s3SchemaVersion": "1.0",
		"configurationId": "tf-s3-lambda-20220411120846560300000001",
		"bucket": {
			"name": "bucket-name",
			"ownerIdentity": {
				"principalId": "SOME_OWNER"
			},
			"arn": "arn:aws:s3:::bucket-name"
		},
		"object": {
			"key": "dir/file.ext",
			"size": 12345,
			"eTag": "6f17b4298e838b30691db31b1d0bc4ec-3",
			"sequencer": "00625807EEBA91FBCA"
		}
	}
}`

func TestMarshall(t *testing.T) {
	loc := time.Location{}
	obj := domain.LambdaRecord{
		EventVersion: "2.1",
		EventSource:  "aws:s3",
		AwsRegion:    "us-west-2",
		EventTime:    domain.JsonTime(time.Date(2022, 04, 14, 11, 39, 29, 346000000, &loc)),
		EventName:    "ObjectCreated:CompleteMultipartUpload",
		UserIdentity: domain.LambdaUserIdentity{
			PrincipalId: "AWS:SOMEPRINCIPAL",
		},
		RequestParameters: domain.LambdaRequestParameters{
			SourceIPAddress: "123.45.67.89",
		},
		ResponseElements: domain.LambdaResponseElements{
			RequestId: "XT6FD2FBQWXM1ABC",
			Id2:       "ab7rhq6747Kpa/aBY60gVUd1kd79J7asNC3RvyN6d77zjzYn+aBnTh5107THtwu/qufcgLisDK+30aErdEbk7Rw7a5EokaBC",
		},
		S3: domain.S3Record{
			S3SchemaVersion: "1.0",
			ConfigurationId: "tf-s3-lambda-20220411120846560300000001",
			Bucket: domain.S3Bucket{
				Name:          "bucket-name",
				OwnerIdentity: domain.S3BucketOwnerIdentity{PrincipalId: "SOME_OWNER"},
				Arn:           "arn:aws:s3:::bucket-name",
			},
			Object: domain.S3Object{
				Key:       "dir/file.ext",
				Size:      12345,
				ETag:      "6f17b4298e838b30691db31b1d0bc4ec-3",
				Sequencer: "00625807EEBA91FBCA",
			},
		},
	}

	bytes, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		t.Fatalf("Unable to marshall: %v", err)
	}

	assert.Equal(t, expected, string(bytes))
}
