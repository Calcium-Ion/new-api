package service

import (
	"encoding/base64"
	"fmt"
)

func parseAudio(audioBase64 string, format string) (duration float64, err error) {
	audioData, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		return 0, fmt.Errorf("base64 decode error: %v", err)
	}

	var samplesCount int
	var sampleRate int

	switch format {
	case "pcm16":
		samplesCount = len(audioData) / 2 // 16位 = 2字节每样本
		sampleRate = 24000                // 24kHz
	case "g711_ulaw", "g711_alaw":
		samplesCount = len(audioData) // 8位 = 1字节每样本
		sampleRate = 8000             // 8kHz
	default:
		samplesCount = len(audioData) // 8位 = 1字节每样本
		sampleRate = 8000             // 8kHz
	}

	duration = float64(samplesCount) / float64(sampleRate)
	return duration, nil
}
