package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var languageKeywords = map[string][]string{
	".py":   {"def", "class"},         // Python
	".js":   {"function", "class"},    // JavaScript
	".go":   {"func", "type"},         // Go
	".java": {"class", "interface"},   // Java
	".cpp":  {"class", "void", "int"}, // C++
}

// GetKeywordsByExtension은 파일 확장자를 기반으로 키워드를 가져온다.
func GetKeywordsByExtension(path string) ([]string, error) {
	extension := filepath.Ext(path)
	if keywords, exists := languageKeywords[extension]; exists {
		return keywords, nil
	}
	return nil, fmt.Errorf("unsupported file extension: %s", extension)
}

// SplitByKeywords는 키워드로 코드를 분리한다.
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

func NormalizeDeclaration(declaration string) string {
	return strings.Join(strings.Fields(declaration), " ")
}

// ExtractDeclaration은 코드 청크에서 함수 또는 클래스의 선언부(헤더)만을 추출하고 정규화한다.
func ExtractDeclaration(chunk string, extension string) string {
	chunk = strings.TrimSpace(chunk)
	if chunk == "" {
		return ""
	}

	var declaration string
	lines := strings.Split(chunk, "\n")
	var declarationParts []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		declarationParts = append(declarationParts, trimmed)

		// Python의 경우 ':'로 끝나는 부분까지가 선언부
		if extension == ".py" && strings.HasSuffix(trimmed, ":") {
			declaration = strings.Join(declarationParts, " ")
			break
		} else if extension == ".go" || extension == ".java" || extension == ".cpp" || extension == ".js" {
			// 다른 언어들은 '{'가 나오는 부분까지가 선언부
			if idx := strings.Index(trimmed, "{"); idx >= 0 {
				declarationParts[len(declarationParts)-1] = trimmed[:idx]
				declaration = strings.Join(declarationParts, " ")
				break
			}
		} else if extension == "" {
			declaration = strings.Join(declarationParts, " ")
			break
		}
	}

	return NormalizeDeclaration(declaration)
}

func HashChunk(chunk string) string {
	hash := sha256.New()
	hash.Write([]byte(chunk))
	return hex.EncodeToString(hash.Sum(nil))
}
