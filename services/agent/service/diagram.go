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

		return result, nil
	}

	return nil, fmt.Errorf("unexpected error in diagram generation")
}

// callOnce는 단일 시도로 다이어그램을 생성합니다
func (agent DiagramAgent) callOnce(code string, purpose string, diagramType DiagramType, attempt int) (*DiagramResult, error) {
	retryNote := ""
	if attempt > 1 {
		retryNote = fmt.Sprintf("\n\nThis is attempt #%d. Please ensure the diagram follows proper Mermaid syntax and includes meaningful content.", attempt)
	}

	prompt := fmt.Sprintf(`Please analyze the following code and create a Mermaid %s diagram.
				Code:
				%s

				Purpose: %s

				Please follow these rules:
				1. Create a %s diagram that clearly visualizes the code structure
				2. Show relationships between components/functions appropriately
				3. Use Mermaid syntax starting with '%s'
				4. Ensure the diagram is syntactically correct and meaningful
				5. Include proper node names and connections%s

				Return ONLY the mermaid diagram code as a string, without any additional formatting or explanation.`,
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
			Description: "Flowchart diagram for visualizing control flow and function calls",
			UseCase:     "Use when code shows primarily function calls and control flow logic",
		},
		{
			Type:        DiagramTypeSequence,
			Description: "Sequence diagram for showing interactions between components over time",
			UseCase:     "Use when code shows sequences of interactions between different components or objects",
		},
		{
			Type:        DiagramTypeClass,
			Description: "Class diagram for showing object-oriented relationships",
			UseCase:     "Use when code defines many classes/structs with inheritance or composition relationships",
		},
		{
			Type:        DiagramTypeER,
			Description: "Entity-Relationship diagram for data modeling",
			UseCase:     "Use when code deals primarily with data relationships and database structures",
		},
		{
			Type:        DiagramTypeComponent,
			Description: "Component diagram for system architecture",
			UseCase:     "Use when code defines a system with multiple distinct components or modules",
		},
		{
			Type:        DiagramTypeState,
			Description: "State diagram for state machines and state transitions",
			UseCase:     "Use when code implements state machines or has clear state transitions",
		},
	}

	optionsJSON, _ := json.Marshal(options)

	prompt := fmt.Sprintf(`Analyze the following code and determine the most appropriate Mermaid diagram type to visualize it:

			Code:
			%s

			Available diagram type options:
			%s

			Based on the code structure, content, and complexity, select the ONE most appropriate diagram type from the options above.

			Return your selection in JSON format with the following structure:
			{
				"selectedType": "the_selected_diagram_type",
				"reason": "explanation of why this type is most suitable for the given code"
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
	case DiagramTypeER:
		return "erDiagram"
	case DiagramTypeComponent:
		return "componentDiagram"
	case DiagramTypeState:
		return "stateDiagram-v2"
	default:
		return "flowchart TD"
	}
}
