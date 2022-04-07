package domain

type FilterRule struct {
	Name  string
	Value string
}

type S3Key struct {
	FilterRules []FilterRule `xml:"FilterRule"`
}

type Filter struct {
	S3Key S3Key
}

type CloudFunctionConfiguration struct {
	Events        []string `xml:"Event"`
	Filter        Filter
	ID            string
	CloudFunction string
}

type NotificationConfiguration struct {
	CloudFunctionConfigurations []CloudFunctionConfiguration `xml:"CloudFunctionConfiguration"`
}
