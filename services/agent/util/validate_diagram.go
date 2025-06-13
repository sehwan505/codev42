package util

import (
	"fmt"
	"regexp"
	"strings"
)

type DiagramType string

const (
	DiagramTypeFlowchart DiagramType = "flowchart"
	DiagramTypeSequence  DiagramType = "sequence"
	DiagramTypeClass     DiagramType = "class"
)

type DiagramValidator struct{}

func NewDiagramValidator() *DiagramValidator {
	return &DiagramValidator{}
}

// 다이어그램 검증 함수 (개선된 버전)
func (validator DiagramValidator) ValidateDiagram(diagram string, diagramType DiagramType) error {
	diagram = strings.TrimSpace(diagram)

	// 기본 검증
	if err := validator.basicValidation(diagram); err != nil {
		return err
	}

	// 접두어 검증
	if err := validator.validatePrefix(diagram, diagramType); err != nil {
		return err
	}

	// 구문 검증
	if err := validator.validateSyntax(diagram, diagramType); err != nil {
		return err
	}

	// 타입별 구체적 검증
	return validator.validateTypeSpecific(diagram, diagramType)
}

// 기본 검증
func (validator DiagramValidator) basicValidation(diagram string) error {
	if diagram == "" {
		return fmt.Errorf("diagram is empty")
	}

	lines := strings.Split(diagram, "\n")
	if len(lines) < 2 {
		return fmt.Errorf("diagram is too short, expected at least 2 lines")
	}

	return nil
}

// 접두어 검증
func (validator DiagramValidator) validatePrefix(diagram string, diagramType DiagramType) error {
	expectedPrefix := getMermaidPrefix(diagramType)
	firstLine := strings.TrimSpace(strings.Split(diagram, "\n")[0])

	if !strings.HasPrefix(firstLine, expectedPrefix) {
		return fmt.Errorf("diagram does not start with expected prefix '%s', got '%s'",
			expectedPrefix, firstLine)
	}

	return nil
}

// 구문 검증 (개선된 버전)
func (validator DiagramValidator) validateSyntax(diagram string, diagramType DiagramType) error {
	// 일반적인 Mermaid 구문 오류 체크
	syntaxErrors := []struct {
		pattern string
		message string
	}{
		{`<<[^>]*>>.*<<`, "nested angle brackets are not allowed"},
		{`\s+<<\s+`, "invalid spacing around angle brackets"},
		{`>>\s+<<`, "invalid annotation sequence"},
		{`<<[^>]*\n[^>]*>>`, "annotations cannot span multiple lines"},
	}

	for _, check := range syntaxErrors {
		if matched, _ := regexp.MatchString(check.pattern, diagram); matched {
			return fmt.Errorf("syntax error: %s", check.message)
		}
	}

	return nil
}

// 타입별 구체적 검증
func (validator DiagramValidator) validateTypeSpecific(diagram string, diagramType DiagramType) error {
	switch diagramType {
	case DiagramTypeFlowchart:
		return validator.validateFlowchart(diagram)
	case DiagramTypeSequence:
		return validator.validateSequence(diagram)
	case DiagramTypeClass:
		return validator.validateClass(diagram)
	default:
		return nil
	}
}

// 플로우차트 검증
func (validator DiagramValidator) validateFlowchart(diagram string) error {
	connectionPatterns := []string{"-->", "---", "-.->", "-.-", "==>", "=="}
	hasConnection := false

	for _, pattern := range connectionPatterns {
		if strings.Contains(diagram, pattern) {
			hasConnection = true
			break
		}
	}

	if !hasConnection {
		return fmt.Errorf("flowchart diagram should contain connections (-->, ---, -.->, etc.)")
	}

	// 노드 정의 검증
	nodePattern := regexp.MustCompile(`\w+\[[^\]]+\]|\w+\([^)]+\)|\w+\{[^}]+\}`)
	if !nodePattern.MatchString(diagram) {
		return fmt.Errorf("flowchart should contain properly defined nodes")
	}

	return nil
}

// 시퀀스 다이어그램 검증
func (validator DiagramValidator) validateSequence(diagram string) error {
	messagePatterns := []string{"->>", "->", "-->>", "-->", "-x", "--x"}
	hasMessage := false

	for _, pattern := range messagePatterns {
		if strings.Contains(diagram, pattern) {
			hasMessage = true
			break
		}
	}

	if !hasMessage {
		return fmt.Errorf("sequence diagram should contain message arrows (->>, ->, -->, etc.)")
	}

	// 참가자 검증
	if !strings.Contains(diagram, "participant") && !regexp.MustCompile(`\w+\s*-`).MatchString(diagram) {
		return fmt.Errorf("sequence diagram should contain participants")
	}

	return nil
}

// 클래스 다이어그램 검증 (개선된 버전)
func (validator DiagramValidator) validateClass(diagram string) error {
	// 클래스 정의 패턴들
	classPatterns := []*regexp.Regexp{
		regexp.MustCompile(`class\s+\w+`),              // class ClassName
		regexp.MustCompile(`\w+\s*:\s*\w+`),            // ClassName : method
		regexp.MustCompile(`\w+\s*\|\s*\w+`),           // ClassName | attribute
		regexp.MustCompile(`\w+\s*{\s*[\w\s:()]+\s*}`), // ClassName { content }
	}

	hasClassDefinition := false
	for _, pattern := range classPatterns {
		if pattern.MatchString(diagram) {
			hasClassDefinition = true
			break
		}
	}

	if !hasClassDefinition {
		return fmt.Errorf("class diagram should contain class definitions")
	}

	// 어노테이션 검증 (에러 원인 해결)
	if err := validator.validateClassAnnotations(diagram); err != nil {
		return err
	}

	// 관계가 없어도 단일 클래스 다이어그램은 유효할 수 있음
	return nil
}

// 클래스 다이어그램 어노테이션 검증 (에러 해결)
func (validator DiagramValidator) validateClassAnnotations(diagram string) error {
	// 잘못된 어노테이션 패턴들
	invalidPatterns := []struct {
		pattern *regexp.Regexp
		message string
	}{
		{
			regexp.MustCompile(`<<\s*\w+\s*>>\s*<<`),
			"consecutive annotations without proper separation",
		},
		{
			regexp.MustCompile(`<<[^>]*\n[^>]*>>`),
			"annotations cannot span multiple lines",
		},
		{
			regexp.MustCompile(`<<\s*>>`),
			"empty annotations are not allowed",
		},
		{
			regexp.MustCompile(`<<[^>]*[^>\w\s][^>]*>>`),
			"annotations should only contain alphanumeric characters and spaces",
		},
	}

	for _, check := range invalidPatterns {
		if check.pattern.MatchString(diagram) {
			return fmt.Errorf("annotation error: %s", check.message)
		}
	}

	// 유효한 어노테이션 형식 확인
	annotationPattern := regexp.MustCompile(`<<\s*\w+\s*>>`)
	annotations := annotationPattern.FindAllString(diagram, -1)

	for _, annotation := range annotations {
		// 어노테이션이 클래스명 뒤에 올바르게 위치하는지 확인
		if !validator.isValidAnnotationPlacement(diagram, annotation) {
			return fmt.Errorf("annotation '%s' is not properly placed", annotation)
		}
	}

	return nil
}

// 어노테이션 배치 검증
func (validator DiagramValidator) isValidAnnotationPlacement(diagram, annotation string) bool {
	lines := strings.Split(diagram, "\n")

	for _, line := range lines {
		if strings.Contains(line, annotation) {
			// 어노테이션이 있는 라인에서 클래스명이나 관계 정의가 있는지 확인
			trimmed := strings.TrimSpace(line)

			// 클래스 정의 라인
			if strings.HasPrefix(trimmed, "class ") {
				return true
			}

			// 관계 정의에서의 어노테이션
			relationshipPattern := regexp.MustCompile(`\w+\s*(--|\.\.|\->|<\|).*<<.*>>`)
			if relationshipPattern.MatchString(trimmed) {
				return true
			}
		}
	}

	return false
}

// Mermaid 접두어 반환
func getMermaidPrefix(diagramType DiagramType) string {
	switch diagramType {
	case DiagramTypeFlowchart:
		return "flowchart"
	case DiagramTypeSequence:
		return "sequenceDiagram"
	case DiagramTypeClass:
		return "classDiagram"
	default:
		return ""
	}
}
