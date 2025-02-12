package service

import (
	"errors"
	"fmt"
	"one-api/dto"
	"one-api/setting"
	"strings"
)

func CheckSensitiveMessages(messages []dto.Message) error {
	for _, message := range messages {
		if len(message.Content) > 0 {
			if message.IsStringContent() {
				stringContent := message.StringContent()
				if ok, words := SensitiveWordContains(stringContent); ok {
					return errors.New("sensitive words: " + strings.Join(words, ","))
				}
			}
		} else {
			arrayContent := message.ParseContent()
			for _, m := range arrayContent {
				if m.Type == "image_url" {
					// TODO: check image url
				} else {
					if ok, words := SensitiveWordContains(m.Text); ok {
						return errors.New("sensitive words: " + strings.Join(words, ","))
					}
				}
			}
		}
	}
	return nil
}

func CheckSensitiveText(text string) error {
	if ok, words := SensitiveWordContains(text); ok {
		return errors.New("sensitive words: " + strings.Join(words, ","))
	}
	return nil
}

func CheckSensitiveInput(input any) error {
	switch v := input.(type) {
	case string:
		return CheckSensitiveText(v)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return CheckSensitiveText(text)
	}
	return CheckSensitiveText(fmt.Sprintf("%v", input))
}

// SensitiveWordContains 是否包含敏感词，返回是否包含敏感词和敏感词列表
func SensitiveWordContains(text string) (bool, []string) {
	if len(setting.SensitiveWords) == 0 {
		return false, nil
	}
	checkText := strings.ToLower(text)
	// 构建一个AC自动机
	m := InitAc()
	hits := m.MultiPatternSearch([]rune(checkText), false)
	if len(hits) > 0 {
		words := make([]string, 0)
		for _, hit := range hits {
			words = append(words, string(hit.Word))
		}
		return true, words
	}
	return false, nil
}

// SensitiveWordReplace 敏感词替换，返回是否包含敏感词和替换后的文本
func SensitiveWordReplace(text string, returnImmediately bool) (bool, []string, string) {
	if len(setting.SensitiveWords) == 0 {
		return false, nil, text
	}
	checkText := strings.ToLower(text)
	m := InitAc()
	hits := m.MultiPatternSearch([]rune(checkText), returnImmediately)
	if len(hits) > 0 {
		words := make([]string, 0)
		for _, hit := range hits {
			pos := hit.Pos
			word := string(hit.Word)
			text = text[:pos] + "**###**" + text[pos+len(word):]
			words = append(words, word)
		}
		return true, words, text
	}
	return false, nil, text
}
