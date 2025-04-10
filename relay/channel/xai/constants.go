package xai

var ModelList = []string{
	// grok-3
	"grok-3-beta", "grok-3-mini-beta",
	// grok-3 mini
	"grok-3-fast-beta", "grok-3-mini-fast-beta",
	// extend grok-3-mini reasoning
	"grok-3-mini-beta-high", "grok-3-mini-beta-low", "grok-3-mini-beta-medium",
	"grok-3-mini-fast-beta-high", "grok-3-mini-fast-beta-low", "grok-3-mini-fast-beta-medium",
	// image model
	"grok-2-image",
	// legacy models
	"grok-2", "grok-2-vision",
	"grok-beta", "grok-vision-beta",
}

var ChannelName = "xai"
