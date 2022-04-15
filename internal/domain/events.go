package domain

const (
	ObjectCreatedEvent = "s3:ObjectCreated"
	ObjectRemovedEvent = "s3:ObjectRemoved"
)

type NotificationEvent struct {
	Bucket   string
	Key      string // S3 Object key
	Event    string // S3 event (i.e. s3:ObjectCreated", "s3:ObjectRemoved", etc.)
	SourceIp string
	Size     int64
}
