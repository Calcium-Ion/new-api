package service

import (
	"bytes"
	"fmt"
	"github.com/anknown/ahocorasick"
	"one-api/constant"
	"strings"
)

// SensitiveWordContains 是否包含敏感词，返回是否包含敏感词和敏感词列表
func SensitiveWordContains(text string) (bool, []string) {
	// 构建一个AC自动机
	m := initAc()
	hits := m.MultiPatternSearch([]rune(text), false)
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
func SensitiveWordReplace(text string) (bool, string) {
	m := initAc()
	hits := m.MultiPatternSearch([]rune(text), false)
	if len(hits) > 0 {
		for _, hit := range hits {
			pos := hit.Pos
			word := string(hit.Word)
			text = text[:pos] + strings.Repeat("*", len(word)) + text[pos+len(word):]
		}
		return true, text
	}
	return false, text
}

func initAc() *goahocorasick.Machine {
	m := new(goahocorasick.Machine)
	dict := readRunes()
	if err := m.Build(dict); err != nil {
		fmt.Println(err)
		return nil
	}
	return m
}

func readRunes() [][]rune {
	var dict [][]rune

	for _, word := range constant.SensitiveWords {
		l := bytes.TrimSpace([]byte(word))
		dict = append(dict, bytes.Runes(l))
	}

	return dict
}
