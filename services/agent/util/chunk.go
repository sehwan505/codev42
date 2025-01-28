package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var languageKeywords = map[string][]string{
	".py":   {"def", "class"},         // Python
	".js":   {"function", "class"},    // JavaScript
	".go":   {"func", "type"},         // Go
	".java": {"class", "interface"},   // Java
	".cpp":  {"class", "void", "int"}, // C++
}

// 파일 확장자를 기반으로 키워드 가져오기
func GetKeywordsByExtension(path string) ([]string, error) {
	extension := filepath.Ext(path)
	if keywords, exists := languageKeywords[extension]; exists {
		return keywords, nil
	}
	return nil, fmt.Errorf("unsupported file extension: %s", extension)
}

// 키워드로 코드를 분리하는 함수
func SplitByKeywords(code string, keywords []string) []string {
	keywordPattern := strings.Join(keywords, "|")
	re := regexp.MustCompile(fmt.Sprintf(`(%s)`, keywordPattern))

	var chunks []string
	lastIndex := 0

	matches := re.FindAllStringIndex(code, -1)
	for _, match := range matches {
		if lastIndex < match[0] {
			chunks = append(chunks, code[lastIndex:match[0]])
		}
		lastIndex = match[0]
	}

	if lastIndex < len(code) {
		chunks = append(chunks, code[lastIndex:])
	}

	return chunks
}

// ExtractName은 코드 청크에서 함수/클래스 이름을 추출
func ExtractName(chunk string, keywords []string) string {
	lines := strings.Split(strings.TrimSpace(chunk), "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := lines[0]
	for _, keyword := range keywords {
		if strings.Contains(firstLine, keyword) {
			parts := strings.Fields(firstLine)
			for i, part := range parts {
				if part == keyword && i+1 < len(parts) {
					name := parts[i+1]
					// 괄호나 기타 기호 제거
					return strings.TrimFunc(name, func(r rune) bool {
						return !unicode.IsLetter(r) && !unicode.IsNumber(r)
					})
				}
			}
		}
	}
	return ""
}

// HashChunk은 코드 청크를 해시
func HashChunk(chunk string) string {
	hash := sha256.New()
	hash.Write([]byte(chunk))
	return hex.EncodeToString(hash.Sum(nil))
}
