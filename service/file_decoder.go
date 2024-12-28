package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"one-api/constant"
	"one-api/dto"
)

var maxFileSize = constant.MaxFileDownloadMB * 1024 * 1024

func GetFileBase64FromUrl(url string) (*dto.LocalFileData, error) {
	resp, err := DoDownloadRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Always use LimitReader to prevent oversized downloads
	fileBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxFileSize+1)))
	if err != nil {
		return nil, err
	}

	// Check actual size after reading
	if len(fileBytes) > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size: %dMB", constant.MaxFileDownloadMB)
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(fileBytes)

	return &dto.LocalFileData{
		Base64Data: base64Data,
		MimeType:   resp.Header.Get("Content-Type"),
		Size:       int64(len(fileBytes)),
	}, nil
}
