package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// OpenAPI 3.0文档生成器
// 这个脚本会读取swagger.json文件并进行简单处理
func main() {
	fmt.Println("正在生成OpenAPI 3.0文档...")

	// 读取swagger.json文件
	docsDir, err := filepath.Abs("docs")
	if err != nil {
		fmt.Printf("获取文档目录失败: %v\n", err)
		return
	}

	swaggerFile := filepath.Join(docsDir, "swagger.json")
	data, err := ioutil.ReadFile(swaggerFile)
	if err != nil {
		fmt.Printf("读取swagger.json失败: %v\n", err)
		return
	}

	// 解析JSON
	var openapiDoc map[string]interface{}
	err = json.Unmarshal(data, &openapiDoc)
	if err != nil {
		fmt.Printf("解析swagger.json失败: %v\n", err)
		return
	}

	// 简单验证必要字段
	if _, ok := openapiDoc["openapi"]; !ok {
		fmt.Println("警告: OpenAPI版本字段缺失")
	}
	if _, ok := openapiDoc["info"]; !ok {
		fmt.Println("警告: info字段缺失")
	}
	if _, ok := openapiDoc["paths"]; !ok {
		fmt.Println("警告: paths字段缺失")
	}
	if _, ok := openapiDoc["components"]; !ok {
		fmt.Println("警告: components字段缺失 (OpenAPI 3.0要求)")
	}

	// 重新写入美化后的JSON
	prettyJSON, err := json.MarshalIndent(openapiDoc, "", "  ")
	if err != nil {
		fmt.Printf("JSON格式化失败: %v\n", err)
		return
	}

	err = ioutil.WriteFile(swaggerFile, prettyJSON, os.ModePerm)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}

	// 复制到web目录
	webSwaggerDir := filepath.Join("web", "public", "swagger")
	err = os.MkdirAll(webSwaggerDir, os.ModePerm)
	if err != nil {
		fmt.Printf("创建web swagger目录失败: %v\n", err)
	} else {
		err = ioutil.WriteFile(filepath.Join(webSwaggerDir, "swagger.json"), prettyJSON, os.ModePerm)
		if err != nil {
			fmt.Printf("复制到web目录失败: %v\n", err)
		}
	}

	fmt.Println("OpenAPI 3.0文档生成完成!")
	fmt.Println("请访问 http://localhost:3000/swagger/index.html 查看API文档")
}
