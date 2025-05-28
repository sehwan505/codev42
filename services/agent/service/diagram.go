package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/openai/openai-go"
)

type DiagramAgent struct {
	Client        *openai.Client
	AnalyserAgent *AnalyserAgent // 코드 세그먼트 분석을 위한 AnalyserAgent 추가
}

func NewDiagramAgent(apiKey string) *DiagramAgent {
	openaiClient := GetClient(apiKey)

	return &DiagramAgent{
		Client:        openaiClient.Client(),
		AnalyserAgent: NewAnalyserAgent(apiKey), // AnalyserAgent 초기화
	}
}

type DiagramType string

const (
	DiagramTypeFlowchart DiagramType = "flowchart"
	DiagramTypeSequence  DiagramType = "sequence"
	DiagramTypeClass     DiagramType = "class"
	DiagramTypeER        DiagramType = "er"
	DiagramTypeComponent DiagramType = "component"
	DiagramTypeState     DiagramType = "stateDiagram"
)

// DiagramResult는 다이어그램 생성 결과를 나타내는 구조체입니다
type DiagramResult struct {
	Diagram string      `json:"diagram"` // Mermaid 다이어그램 코드
	Type    DiagramType `json:"type"`    // 다이어그램 타입
}

// DiagramTypeOption은 다이어그램 타입 선택 옵션을 나타내는 구조체입니다
type DiagramTypeOption struct {
	Type        DiagramType `json:"type"`        // 다이어그램 타입
	Description string      `json:"description"` // 타입 설명
	UseCase     string      `json:"useCase"`     // 사용 사례
}

type DiagramTypeSelectionResult struct {
	SelectedType []DiagramType `json:"list of selectedType"` // 선택된 다이어그램 타입
}

// 타입 검증 메서드
func (d DiagramType) IsValid() bool {
	switch d {
	case DiagramTypeFlowchart, DiagramTypeSequence, DiagramTypeClass,
		DiagramTypeER, DiagramTypeComponent, DiagramTypeState:
		return true
	}
	return false
}

// 다이어그램 검증 함수
func (agent DiagramAgent) validateDiagram(diagram string, diagramType DiagramType) error {
	// 기본 검증: 다이어그램이 비어있지 않은지 확인
	if strings.TrimSpace(diagram) == "" {
		return fmt.Errorf("diagram is empty")
	}

	// Mermaid 접두어 검증
	expectedPrefix := getMermaidPrefix(diagramType)
	if !strings.HasPrefix(strings.TrimSpace(diagram), expectedPrefix) {
		return fmt.Errorf("diagram does not start with expected prefix '%s'", expectedPrefix)
	}

	// 기본적인 Mermaid 구문 검증
	lines := strings.Split(diagram, "\n")
	if len(lines) < 2 {
		return fmt.Errorf("diagram is too short, expected at least 2 lines")
	}

	// 다이어그램 타입별 추가 검증
	switch diagramType {
	case DiagramTypeFlowchart:
		if !strings.Contains(diagram, "-->") && !strings.Contains(diagram, "---") {
			return fmt.Errorf("flowchart diagram should contain connections (-->, ---)")
		}
	case DiagramTypeSequence:
		if !strings.Contains(diagram, "->>") && !strings.Contains(diagram, "->") {
			return fmt.Errorf("sequence diagram should contain message arrows (->>, ->)")
		}
	case DiagramTypeClass:
		if !strings.Contains(diagram, "class ") && !strings.Contains(diagram, ":") {
			return fmt.Errorf("class diagram should contain class definitions")
		}
	}

	return nil
}

// call은 단일 다이어그램 생성을 처리하는 내부 메서드입니다 (재시도 로직 포함)
func (agent DiagramAgent) call(code string, purpose string, diagramType DiagramType) (*DiagramResult, error) {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := agent.callOnce(code, purpose, diagramType, attempt)
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to generate diagram after %d attempts: %v", maxRetries, err)
			}
			fmt.Printf("Attempt %d failed, retrying: %v\n", attempt, err)
			continue
		}

		// 다이어그램 검증
		if err := agent.validateDiagram(result.Diagram, diagramType); err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("diagram validation failed after %d attempts: %v", maxRetries, err)
			}
			fmt.Printf("Attempt %d validation failed, retrying: %v\n", attempt, err)
			continue
		}
		fmt.Println("result: ", result, "attempt: ", attempt)
		return result, nil
	}

	return nil, fmt.Errorf("unexpected error in diagram generation")
}

// callOnce는 단일 시도로 다이어그램을 생성합니다
func (agent DiagramAgent) callOnce(code string, purpose string, diagramType DiagramType, attempt int) (*DiagramResult, error) {
	retryNote := ""
	if attempt > 1 {
		retryNote = fmt.Sprintf("\n\n이것은 %d번째 시도입니다. 다이어그램이 적절한 Mermaid 문법을 따르고 의미있는 내용을 포함하도록 해주세요.", attempt)
	}

	prompt := fmt.Sprintf(`다음 코드를 분석하여 Mermaid %s 다이어그램을 생성해주세요.
				코드:
				%s

				목적: %s

				다음 규칙을 따라주세요:
				1. 코드 구조를 명확하게 시각화하는 %s 다이어그램을 생성해주세요
				2. 컴포넌트/함수 간의 관계를 적절하게 보여주세요
				3. '%s'로 시작하는 Mermaid 문법을 사용해주세요
				4. 다이어그램이 문법적으로 올바르고 의미있는지 확인해주세요
				5. 적절한 노드 이름과 연결을 포함해주세요
				6. 모든 노드 이름과 라벨은 한국어로 작성해주세요%s

				mermaid 다이어그램 코드만을 문자열로 반환해주세요. 추가적인 형식이나 설명은 포함하지 마세요.`,
		diagramType, code, purpose, diagramType, getMermaidPrefix(diagramType), retryNote)

	// 간단한 스키마 정의 - 다이어그램 코드만 받기
	simpleDiagramResultSchema := GenerateImplementResultSchema[DiagramResult]()

	chat, err := agent.Client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("diagram_result"),
					Description: openai.F("mermaid diagram code with type"),
					Schema:      openai.F(simpleDiagramResultSchema),
					Strict:      openai.Bool(true),
				}),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o2024_11_20),
	})

	if err != nil {
		return nil, err
	}

	var simpleDiagramResult DiagramResult
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &simpleDiagramResult)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// 최종 결과 구성
	diagramResult := &DiagramResult{
		Diagram: simpleDiagramResult.Diagram,
		Type:    diagramType, // 요청한 타입으로 설정
	}

	return diagramResult, nil
}

// ImplementDiagrams는 여러 다이어그램 요청을 병렬로 처리합니다
func (agent DiagramAgent) ImplementDiagrams(code string, purpose string, diagramTypes []DiagramType) ([]*DiagramResult, error) {
	var wg sync.WaitGroup
	resultChan := make(chan *DiagramResult, len(diagramTypes))
	errorChan := make(chan error, len(diagramTypes))

	for _, diagramType := range diagramTypes {
		wg.Add(1)
		go func(diagramType DiagramType) {
			defer wg.Done()
			fmt.Printf("Processing diagram: %s\n", diagramType)

			result, err := agent.call(code, purpose, diagramType)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- result
		}(diagramType)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	var results []*DiagramResult
	for result := range resultChan {
		results = append(results, result)
	}

	if len(errorChan) > 0 {
		var errors []string
		for err := range errorChan {
			errors = append(errors, err.Error())
		}
		return results, fmt.Errorf("some diagrams failed to generate: %v", errors)
	}

	return results, nil
}

func (agent DiagramAgent) SelectOptimalDiagramType(code string) (DiagramTypeSelectionResult, error) {
	// 다이어그램 타입 옵션들을 구조체로 정의
	options := []DiagramTypeOption{
		{
			Type:        DiagramTypeFlowchart,
			Description: "제어 흐름과 함수 호출을 시각화하는 플로우차트 다이어그램",
			UseCase:     "코드가 주로 함수 호출과 제어 흐름 로직을 보여줄 때 사용",
		},
		{
			Type:        DiagramTypeSequence,
			Description: "시간에 따른 컴포넌트 간 상호작용을 보여주는 시퀀스 다이어그램",
			UseCase:     "코드가 서로 다른 컴포넌트나 객체 간의 상호작용 시퀀스를 보여줄 때 사용",
		},
		{
			Type:        DiagramTypeClass,
			Description: "객체지향 관계를 보여주는 클래스 다이어그램",
			UseCase:     "코드가 상속이나 구성 관계를 가진 많은 클래스/구조체를 정의할 때 사용",
		},
		{
			Type:        DiagramTypeState,
			Description: "상태 머신과 상태 전환을 위한 상태 다이어그램",
			UseCase:     "코드가 상태 머신을 구현하거나 명확한 상태 전환이 있을 때 사용",
		},
	}

	optionsJSON, _ := json.Marshal(options)

	prompt := fmt.Sprintf(`다음 코드를 분석하여 시각화하기에 가장 적절한 Mermaid 다이어그램 타입을 결정해주세요:

			코드:
			%s

			사용 가능한 다이어그램 타입 옵션:
			%s

			코드 구조, 내용, 복잡성을 바탕으로 위 옵션 중에서 가장 적절한 두 개의 다이어그램 타입을 선택해주세요.

			다음 JSON 형식으로 선택 결과를 반환해주세요:
			{
				"selectedType": "선택된_다이어그램_타입",
			}`, code, string(optionsJSON))

	var selectionSchema = GenerateImplementResultSchema[DiagramTypeSelectionResult]()
	chat, err := agent.Client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("diagram_type_selection"),
					Description: openai.F("Selected diagram type with reasoning"),
					Schema:      openai.F(selectionSchema),
					Strict:      openai.Bool(true),
				}),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o2024_11_20),
	})

	if err != nil {
		return DiagramTypeSelectionResult{}, err
	}

	var selection DiagramTypeSelectionResult
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &selection)
	if err != nil {
		return DiagramTypeSelectionResult{}, err
	}

	for _, selectedType := range selection.SelectedType {
		if !selectedType.IsValid() {
			return DiagramTypeSelectionResult{}, fmt.Errorf("invalid diagram type")
		}
	}
	return selection, nil
}

// getMermaidPrefix는 다이어그램 타입에 따른 Mermaid 접두어를 반환합니다
func getMermaidPrefix(diagramType DiagramType) string {
	switch diagramType {
	case DiagramTypeFlowchart:
		return "flowchart TD"
	case DiagramTypeSequence:
		return "sequenceDiagram"
	case DiagramTypeClass:
		return "classDiagram"
	case DiagramTypeState:
		return "stateDiagram-v2"
	default:
		return "flowchart TD"
	}
}
