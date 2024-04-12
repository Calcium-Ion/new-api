package gemini

const (
	GeminiVisionMaxImageNum = 16
)

var ModelList = []string{
	"gemini-pro", "gemini-1.0-pro-001", "gemini-1.5-pro",
	"gemini-pro-vision", "gemini-1.0-pro-vision-001",

	"gemini-1.0-pro",
	"gemini-1.0-pro-latest",
	"gemini-1.0-pro-vision-latest",
	"gemini-1.5-pro-latest",

	"text-embedding-004",
	"embedding-001",
	"embedding-gecko-001",
}

var ChannelName = "google gemini"
