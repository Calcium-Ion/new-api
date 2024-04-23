package aws

var awsModelIDMap = map[string]string{
	"claude-instant-1.2":       "anthropic.claude-instant-v1",
	"claude-2.0":               "anthropic.claude-v2",
	"claude-2.1":               "anthropic.claude-v2:1",
	"claude-3-sonnet-20240229": "anthropic.claude-3-sonnet-20240229-v1:0",
	"claude-3-opus-20240229":   "anthropic.claude-3-opus-20240229-v1:0",
	"claude-3-haiku-20240307":  "anthropic.claude-3-haiku-20240307-v1:0",
}

var ChannelName = "aws"
