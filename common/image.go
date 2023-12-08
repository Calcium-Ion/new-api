package common

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/chai2010/webp"
	"image"
	"io"
	"net/http"
	"strings"
)

func DecodeBase64ImageData(base64String string) (image.Config, error) {
	// 去除base64数据的URL前缀（如果有）
	if idx := strings.Index(base64String, ","); idx != -1 {
		base64String = base64String[idx+1:]
	}

	// 将base64字符串解码为字节切片
	decodedData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		fmt.Println("Error: Failed to decode base64 string")
		return image.Config{}, err
	}

	// 创建一个bytes.Buffer用于存储解码后的数据
	reader := bytes.NewReader(decodedData)
	config, err := getImageConfig(reader)
	return config, err
}

func DecodeUrlImageData(imageUrl string) (image.Config, error) {
	response, err := http.Get(imageUrl)
	if err != nil {
		SysLog(fmt.Sprintf("fail to get image from url: %s", err.Error()))
		return image.Config{}, err
	}

	// 限制读取的字节数，防止下载整个图片
	limitReader := io.LimitReader(response.Body, 1024*20)
	//data, err := io.ReadAll(limitReader)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("%x", data)
	config, err := getImageConfig(limitReader)
	response.Body.Close()
	return config, err
}

func getImageConfig(reader io.Reader) (image.Config, error) {
	// 读取图片的头部信息来获取图片尺寸
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		err = errors.New(fmt.Sprintf("fail to decode image config(gif, jpg, png): %s", err.Error()))
		SysLog(err.Error())
		config, err = webp.DecodeConfig(reader)
		if err != nil {
			err = errors.New(fmt.Sprintf("fail to decode image config(webp): %s", err.Error()))
			SysLog(err.Error())
		}
	}
	if err != nil {
		return image.Config{}, err
	}
	return config, nil
}
