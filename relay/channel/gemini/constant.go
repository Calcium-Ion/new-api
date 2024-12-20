package gemini

const (
	GeminiVisionMaxImageNum = 16
)

var ModelList = []string{
	"gemini-1.5-pro-latest", "gemini-1.5-flash-latest",
	// old version
	"gemini-1.5-pro-exp-0827", "gemini-1.5-flash-exp-0827",
	// exp
	"gemini-exp-1114", "gemini-exp-1121", "gemini-exp-1206",
	// flash exp
	"gemini-2.0-flash-exp",
	// thinking exp
	"gemini-2.0-flash-thinking-exp",
	"gemini-2.0-flash-thinking-exp-1219",
}

var ChannelName = "google gemini"
