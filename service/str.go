package service

import (
	"bytes"
	"fmt"
	goahocorasick "github.com/anknown/ahocorasick"
	"one-api/setting"
	"strings"
)

func SundaySearch(text string, pattern string) bool {
	// 计算偏移表
	offset := make(map[rune]int)
	for i, c := range pattern {
		offset[c] = len(pattern) - i
	}

	// 文本串长度和模式串长度
	n, m := len(text), len(pattern)

	// 主循环，i表示当前对齐的文本串位置
	for i := 0; i <= n-m; {
		// 检查子串
		j := 0
		for j < m && text[i+j] == pattern[j] {
			j++
		}
		// 如果完全匹配，返回匹配位置
		if j == m {
			return true
		}

		// 如果还有剩余字符，则检查下一位字符在偏移表中的值
		if i+m < n {
			next := rune(text[i+m])
			if val, ok := offset[next]; ok {
				i += val // 存在于偏移表中，进行跳跃
			} else {
				i += len(pattern) + 1 // 不存在于偏移表中，跳过整个模式串长度
			}
		} else {
			break
		}
	}
	return false // 如果没有找到匹配，返回-1
}

func RemoveDuplicate(s []string) []string {
	result := make([]string, 0, len(s))
	temp := map[string]struct{}{}
	for _, item := range s {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func InitAc() *goahocorasick.Machine {
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

	for _, word := range setting.SensitiveWords {
		word = strings.ToLower(word)
		l := bytes.TrimSpace([]byte(word))
		dict = append(dict, bytes.Runes(l))
	}

	return dict
}
