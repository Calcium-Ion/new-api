package gemini

type GeminiChatRequest struct {
	Contents           []GeminiChatContent        `json:"contents"`
	SafetySettings     []GeminiChatSafetySettings `json:"safety_settings,omitempty"`
	GenerationConfig   GeminiChatGenerationConfig `json:"generation_config,omitempty"`
	Tools              []GeminiChatTool           `json:"tools,omitempty"`
	SystemInstructions *GeminiChatContent         `json:"system_instruction,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type FunctionCall struct {
	FunctionName string `json:"name"`
	Arguments    any    `json:"args"`
}

type GeminiFunctionResponseContent struct {
	Name    string `json:"name"`
	Content any    `json:"content"`
}

type FunctionResponse struct {
	Name     string                        `json:"name"`
	Response GeminiFunctionResponseContent `json:"response"`
}

type GeminiPartExecutableCode struct {
	Language string `json:"language,omitempty"`
	Code     string `json:"code,omitempty"`
}

type GeminiPartCodeExecutionResult struct {
	Outcome string `json:"outcome,omitempty"`
	Output  string `json:"output,omitempty"`
}

type GeminiFileData struct {
	MimeType string `json:"mimeType,omitempty"`
	FileUri  string `json:"fileUri,omitempty"`
}

type GeminiPart struct {
	Text                string                         `json:"text,omitempty"`
	InlineData          *GeminiInlineData              `json:"inlineData,omitempty"`
	FunctionCall        *FunctionCall                  `json:"functionCall,omitempty"`
	FunctionResponse    *FunctionResponse              `json:"functionResponse,omitempty"`
	FileData            *GeminiFileData                `json:"fileData,omitempty"`
	ExecutableCode      *GeminiPartExecutableCode      `json:"executableCode,omitempty"`
	CodeExecutionResult *GeminiPartCodeExecutionResult `json:"codeExecutionResult,omitempty"`
}

type GeminiChatContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiChatSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiChatTool struct {
	GoogleSearch          any `json:"googleSearch,omitempty"`
	GoogleSearchRetrieval any `json:"googleSearchRetrieval,omitempty"`
	CodeExecution         any `json:"codeExecution,omitempty"`
	FunctionDeclarations  any `json:"functionDeclarations,omitempty"`
}

type GeminiChatGenerationConfig struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             float64  `json:"topP,omitempty"`
	TopK             float64  `json:"topK,omitempty"`
	MaxOutputTokens  uint     `json:"maxOutputTokens,omitempty"`
	CandidateCount   int      `json:"candidateCount,omitempty"`
	StopSequences    []string `json:"stopSequences,omitempty"`
	ResponseMimeType string   `json:"responseMimeType,omitempty"`
	ResponseSchema   any      `json:"responseSchema,omitempty"`
	Seed             int64    `json:"seed,omitempty"`
}

type GeminiChatCandidate struct {
	Content       GeminiChatContent        `json:"content"`
	FinishReason  *string                  `json:"finishReason"`
	Index         int64                    `json:"index"`
	SafetyRatings []GeminiChatSafetyRating `json:"safetyRatings"`
}

type GeminiChatSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type GeminiChatPromptFeedback struct {
	SafetyRatings []GeminiChatSafetyRating `json:"safetyRatings"`
}

type GeminiChatResponse struct {
	Candidates     []GeminiChatCandidate    `json:"candidates"`
	PromptFeedback GeminiChatPromptFeedback `json:"promptFeedback"`
	UsageMetadata  GeminiUsageMetadata      `json:"usageMetadata"`
}

type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}
