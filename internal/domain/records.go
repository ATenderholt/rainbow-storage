package domain

import "time"

type S3Object struct {
	Key       string `json:"key"`
	Size      int    `json:"size"`
	ETag      string `json:"eTag"`
	Sequencer string `json:"sequencer"`
}

type S3BucketOwnerIdentity struct {
	PrincipalId string `json:"principalId"`
}

type S3Bucket struct {
	Name          string                `json:"name"`
	OwnerIdentity S3BucketOwnerIdentity `json:"ownerIdentity"`
	Arn           string                `json:"arn"`
}

type S3Record struct {
	S3SchemaVersion string   `json:"s3SchemaVersion"`
	ConfigurationId string   `json:"configurationId"`
	Bucket          S3Bucket `json:"bucket"`
	Object          S3Object `json:"object"`
}

type LambdaResponseElements struct {
	RequestId string `json:"x-amz-request-id"`
	Id2       string `json:"x-amz-id-2"`
}

type LambdaRequestParameters struct {
	SourceIPAddress string `json:"sourceIPAddress"`
}

type LambdaUserIdentity struct {
	PrincipalId string `json:"principalId"`
}

type JsonTime time.Time

const timeFormat = "2006-01-02T15:04:05.999Z"

func (t JsonTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Time(t).Format(timeFormat) + "\""), nil
}

func (t *JsonTime) UnmarshalJSON(bytes []byte) error {
	newTime, err := time.Parse(timeFormat, string(bytes))
	if err != nil {
		return err
	}

	*t = JsonTime(newTime)
	return nil
}

type LambdaRecord struct {
	EventVersion      string                  `json:"eventVersion"`
	EventSource       string                  `json:"eventSource"`
	AwsRegion         string                  `json:"awsRegion"`
	EventTime         JsonTime                `json:"eventTime"`
	EventName         string                  `json:"eventName"`
	UserIdentity      LambdaUserIdentity      `json:"userIdentity"`
	RequestParameters LambdaRequestParameters `json:"requestParameters"`
	ResponseElements  LambdaResponseElements  `json:"responseElements"`
	S3                S3Record                `json:"s3"`
}
