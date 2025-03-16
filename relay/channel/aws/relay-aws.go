package aws

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/relay/channel/claude"
	relaycommon "one-api/relay/common"
	"one-api/service"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func newAwsClient(c *gin.Context, info *relaycommon.RelayInfo) (*bedrockruntime.Client, error) {
	awsSecret := strings.Split(info.ApiKey, "|")
	if len(awsSecret) != 3 {
		return nil, errors.New("invalid aws secret key")
	}
	ak := awsSecret[0]
	sk := awsSecret[1]
	region := awsSecret[2]
	client := bedrockruntime.New(bedrockruntime.Options{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(ak, sk, "")),
	})

	return client, nil
}

func wrapErr(err error) *dto.OpenAIErrorWithStatusCode {
	return &dto.OpenAIErrorWithStatusCode{
		StatusCode: http.StatusInternalServerError,
		Error: dto.OpenAIError{
			Message: fmt.Sprintf("%s", err.Error()),
		},
	}
}

func awsModelID(requestModel string) (string, error) {
	if awsModelID, ok := awsModelIDMap[requestModel]; ok {
		return awsModelID, nil
	}

	return requestModel, nil
}

func awsHandler(c *gin.Context, info *relaycommon.RelayInfo, requestMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	awsCli, err := newAwsClient(c, info)
	if err != nil {
		return wrapErr(errors.Wrap(err, "newAwsClient")), nil
	}

	awsModelId, err := awsModelID(c.GetString("request_model"))
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsReq := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	claudeReq_, ok := c.Get("converted_request")
	if !ok {
		return wrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*dto.ClaudeRequest)
	awsClaudeReq := copyRequest(claudeReq)
	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModel(c.Request.Context(), awsReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModel")), nil
	}

	claudeResponse := new(dto.ClaudeResponse)
	err = json.Unmarshal(awsResp.Body, claudeResponse)
	if err != nil {
		return wrapErr(errors.Wrap(err, "unmarshal response")), nil
	}

	openaiResp := claude.ResponseClaude2OpenAI(requestMode, claudeResponse)
	usage := dto.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
	}
	openaiResp.Usage = usage

	c.JSON(http.StatusOK, openaiResp)
	return nil, &usage
}

func awsStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, requestMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	awsCli, err := newAwsClient(c, info)
	if err != nil {
		return wrapErr(errors.Wrap(err, "newAwsClient")), nil
	}

	awsModelId, err := awsModelID(c.GetString("request_model"))
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsReq := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	claudeReq_, ok := c.Get("converted_request")
	if !ok {
		return wrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*dto.ClaudeRequest)

	awsClaudeReq := copyRequest(claudeReq)
	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModelWithResponseStream(c.Request.Context(), awsReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModelWithResponseStream")), nil
	}
	stream := awsResp.GetStream()
	defer stream.Close()

	claudeInfo := &claude.ClaudeResponseInfo{
		ResponseId:   fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Created:      common.GetTimestamp(),
		Model:        info.UpstreamModelName,
		ResponseText: strings.Builder{},
		Usage:        &dto.Usage{},
	}

	for event := range stream.Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			info.SetFirstResponseTime()
			claude.HandleResponseData(c, info, claudeInfo, string(v.Value.Bytes), RequestModeMessage)
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
			return wrapErr(errors.New("unknown response type")), nil
		default:
			fmt.Println("union is nil or unknown type")
			return wrapErr(errors.New("nil or unknown response type")), nil
		}
	}

	claude.HandleFinalResponse(c, info, claudeInfo, RequestModeMessage)

	if resp != nil {
		err = resp.Body.Close()
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "close_response_body_failed", http.StatusInternalServerError), nil
		}
	}
	return nil, claudeInfo.Usage
}
