package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/image/webp"
	"image"
	"io"
	"one-api/common"
	"strings"
)

func DecodeBase64ImageData(base64String string) (image.Config, string, string, error) {
	// 去除base64数据的URL前缀（如果有）
	if idx := strings.Index(base64String, ","); idx != -1 {
		base64String = base64String[idx+1:]
	}

	// 将base64字符串解码为字节切片
	decodedData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		fmt.Println("Error: Failed to decode base64 string")
		return image.Config{}, "", "", err
	}

	// 创建一个bytes.Buffer用于存储解码后的数据
	reader := bytes.NewReader(decodedData)
	config, format, err := getImageConfig(reader)
	return config, format, base64String, err
}

// GetImageFromUrl 获取图片的类型和base64编码的数据
func GetImageFromUrl(url string) (mimeType string, data string, err error) {
	resp, err := DoImageRequest(url)
	if err != nil {
		return
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/") {
		return
	}
	defer resp.Body.Close()
	buffer := bytes.NewBuffer(nil)
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return
	}
	mimeType = resp.Header.Get("Content-Type")
	data = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return
}

func DecodeUrlImageData(imageUrl string) (image.Config, string, error) {
	response, err := DoImageRequest(imageUrl)
	if err != nil {
		common.SysLog(fmt.Sprintf("fail to get image from url: %s", err.Error()))
		return image.Config{}, "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("fail to get image from url: %s", response.Status))
		return image.Config{}, "", err
	}

	var readData []byte
	for _, limit := range []int64{1024 * 8, 1024 * 24, 1024 * 64} {
		common.SysLog(fmt.Sprintf("try to decode image config with limit: %d", limit))

		// 从response.Body读取更多的数据直到达到当前的限制
		additionalData := make([]byte, limit-int64(len(readData)))
		n, _ := io.ReadFull(response.Body, additionalData)
		readData = append(readData, additionalData[:n]...)

		// 使用io.MultiReader组合已经读取的数据和response.Body
		limitReader := io.MultiReader(bytes.NewReader(readData), response.Body)

		var config image.Config
		var format string
		config, format, err = getImageConfig(limitReader)
		if err == nil {
			return config, format, nil
		}
	}

	return image.Config{}, "", err // 返回最后一个错误
}

func getImageConfig(reader io.Reader) (image.Config, string, error) {
	// 读取图片的头部信息来获取图片尺寸
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		err = errors.New(fmt.Sprintf("fail to decode image config(gif, jpg, png): %s", err.Error()))
		common.SysLog(err.Error())
		config, err = webp.DecodeConfig(reader)
		if err != nil {
			err = errors.New(fmt.Sprintf("fail to decode image config(webp): %s", err.Error()))
			common.SysLog(err.Error())
		}
		format = "webp"
	}
	if err != nil {
		return image.Config{}, "", err
	}
	return config, format, nil
}
