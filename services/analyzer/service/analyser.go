package service

import (
	"codev42-analyzer/client"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

type AnalyserAgent struct {
	Client *openai.Client
}

func NewAnalyserAgent(apiKey string) *AnalyserAgent {
	openaiClient := client.GetClient(apiKey)

	return &AnalyserAgent{
		Client: openaiClient.Client(),
	}
}

type CombinedResult struct {
	Code string `json:"code" jsonschema_description:"the result of combining the codes"`
}

type ImplementResult struct {
	Code string `json:"code" jsonschema_description:"the result of implementing the function or class"`
}

func GenerateImplementResultSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func (agent AnalyserAgent) call(codes []string, purpose string) (*CombinedResult, error) {
	prompt := "목적: " + purpose + "\n\n"
	for i, code := range codes {
		prompt += fmt.Sprintf("코드 %d:\n```\n%s\n```\n\n", i+1, code)
	}

	prompt += "목적에 맞는 효율적인 하나의 코드로 결합해주세요. 최대한 기존 함수나 클래스의 개수를 바꾸지 말아주세요. 모든 설명과 주석은 한국어로 작성해주세요."

	var combinedResultSchema = GenerateImplementResultSchema[CombinedResult]()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("combined_result"),
		Description: openai.F("the result of analyzing and combining the codes"),
		Schema:      openai.F(combinedResultSchema),
		Strict:      openai.Bool(true),
	}

	chat, err := agent.Client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o2024_11_20),
	})

	if err != nil {
		return nil, err
	}

	combinedResult := &CombinedResult{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), combinedResult)
	if err != nil {
		return nil, err
	}
	return combinedResult, nil
}

func (agent AnalyserAgent) CombineImplementation(implementResults []*ImplementResult, purpose string) (*CombinedResult, error) {
	var codes []string
	for _, result := range implementResults {
		if strings.TrimSpace(result.Code) != "" {
			codes = append(codes, result.Code)
		}
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("there is no code")
	}

	return agent.call(codes, purpose)
}

type CodeSegment struct {
	StartLine   int    `json:"startLine"`   // 시작 라인
	EndLine     int    `json:"endLine"`     // 끝 라인
	Explanation string `json:"explanation"` // 세그먼트 설명
}

type CodeSegmentAnalysisResult struct {
	CodeSegments []CodeSegment `json:"codeSegments"` // 코드 세그먼트 설명
}

// AnalyzeCodeSegments 코드를 분석하여 중요한 세그먼트들을 식별하고 설명
func (agent AnalyserAgent) AnalyzeCodeSegments(code, language string) ([]CodeSegment, error) {
	// 코드에 줄 번호 추가
	lines := strings.Split(code, "\n")
	numberedCode := ""
	for i, line := range lines {
		numberedCode += fmt.Sprintf("%4d | %s\n", i, line)
	}

	prompt := fmt.Sprintf(`다음 %s 코드를 분석하여 중요한 세그먼트들을 한국어로 설명해주세요:

코드:
%s

코드의 중요한 세그먼트들을 식별하고 한국어로 설명해주세요:
- 각 중요한 세그먼트에 대해 라인 번호 범위를 식별해주세요 (예: 12-30번째 줄)
- 줄 번호는 0부터 시작해주세요.
- 각 세그먼트가 무엇을 하는지와 그 목적을 한국어로 설명해주세요
- 핵심 로직, 함수, 또는 구조적 요소에 집중해주세요

## 설명 대상
- 함수와 메서드의 목적과 역할
- 복잡한 논리 구조나 알고리즘
- 중요 변수와 데이터 구조의 용도
- 예외 처리와 조건문의 의미
- 코드의 전체적인 흐름을 이해하는 데 도움이 되는 정보

## 추가 지침
- 설명은 간결하면서도 명확하게 작성해주세요
- 특히 복잡한 로직이나 이해하기 어려운 부분에 대해 자세히 설명해주세요
- 전체 코드의 흐름과 구조를 이해할 수 있도록 도와주세요
- 언어나 프레임워크 특정 기능에 대해서는 필요할 경우 추가 설명을 제공해주세요
`, language, numberedCode)
	fmt.Println(prompt)

	var segmentResultSchema = GenerateImplementResultSchema[CodeSegmentAnalysisResult]()

	chat, err := agent.Client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("code_segment_analysis"),
					Description: openai.F("analysis of important code segments with explanations"),
					Schema:      openai.F(segmentResultSchema),
					Strict:      openai.Bool(true),
				}),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o2024_11_20),
	})

	if err != nil {
		return nil, err
	}

	var result CodeSegmentAnalysisResult
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code segment analysis result: %v", err)
	}

	return result.CodeSegments, nil
}
