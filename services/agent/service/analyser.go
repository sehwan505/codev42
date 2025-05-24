package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

type AnalyserAgent struct {
	Client *openai.Client
}

func NewAnalyserAgent(apiKey string) *AnalyserAgent {
	openaiClient := GetClient(apiKey)

	return &AnalyserAgent{
		Client: openaiClient.Client(),
	}
}

type CombinedResult struct {
	Code string `json:"code" jsonschema_description:"the result of combining the codes"`
}

func (agent AnalyserAgent) call(codes []string, purpose string) (*CombinedResult, error) {
	prompt := "purpose: " + purpose + "\n\n"
	prompt += "Please analyze the following codes and combine them into one code that is efficient and meets the purpose:\n\n"

	for i, code := range codes {
		prompt += fmt.Sprintf("코드 %d:\n```\n%s\n```\n\n", i+1, code)
	}

	prompt += "Please combine them into one efficient code that meets the purpose."

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
	Description string `json:"description"` // 세그먼트 설명
}

// CodeSegmentAnalysisResult는 코드 세그먼트 분석 결과를 나타내는 구조체입니다
type CodeSegmentAnalysisResult struct {
	CodeSegments []CodeSegment `json:"codeSegments"` // 코드 세그먼트 설명
}

// AnalyzeCodeSegments는 코드를 분석하여 중요한 세그먼트들을 식별하고 설명합니다
func (agent AnalyserAgent) AnalyzeCodeSegments(code, language string) ([]CodeSegment, error) {
	prompt := fmt.Sprintf(`Please analyze the following %s code and identify important segments:

Code:
%s

Please identify important segments of the code and provide explanations:
- For each important segment, identify the line number range (e.g., lines 12-30)
- Explain what each segment does and its purpose
- Focus on key logic, functions, or structural elements

Return the result in JSON format with the following structure:
{
	"codeSegments": [
		{
			"startLine": 12,
			"endLine": 30,
			"description": "This segment implements the parsing logic for..."
		}
	]
}`, language, code)

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
