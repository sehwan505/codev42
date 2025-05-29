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
	case DiagramTypeFlowchart, DiagramTypeSequence, DiagramTypeClass:
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
	promptTemplate := getDiagramPrompt(code, purpose, diagramType)
	prompt := fmt.Sprintf(`코드:
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
		code, purpose, diagramType, getMermaidPrefix(diagramType), retryNote)
	prompt = promptTemplate + "\n" + prompt
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

// GenerateClassDiagram은 클래스 다이어그램을 생성합니다
func (agent DiagramAgent) GenerateClassDiagram(code string, purpose string) (*DiagramResult, error) {
	return agent.call(code, purpose, DiagramTypeClass)
}

// GenerateSequenceDiagram은 시퀀스 다이어그램을 생성합니다
func (agent DiagramAgent) GenerateSequenceDiagram(code string, purpose string) (*DiagramResult, error) {
	return agent.call(code, purpose, DiagramTypeSequence)
}

// GenerateFlowchartDiagram은 플로우차트 다이어그램을 생성합니다
func (agent DiagramAgent) GenerateFlowchartDiagram(code string, purpose string) (*DiagramResult, error) {
	return agent.call(code, purpose, DiagramTypeFlowchart)
}

// ImplementDiagrams는 세 가지 다이어그램을 병렬로 생성합니다
func (agent DiagramAgent) ImplementDiagrams(code string, purpose string) ([]*DiagramResult, error) {
	var wg sync.WaitGroup
	resultChan := make(chan *DiagramResult, 3)
	errorChan := make(chan error, 3)

	// 클래스 다이어그램 생성
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := agent.GenerateClassDiagram(code, purpose)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// 시퀀스 다이어그램 생성
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := agent.GenerateSequenceDiagram(code, purpose)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

	// 플로우차트 다이어그램 생성
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := agent.GenerateFlowchartDiagram(code, purpose)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- result
	}()

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

// getMermaidPrefix는 다이어그램 타입에 따른 Mermaid 접두어를 반환합니다
func getMermaidPrefix(diagramType DiagramType) string {
	switch diagramType {
	case DiagramTypeFlowchart:
		return "flowchart TD"
	case DiagramTypeSequence:
		return "sequenceDiagram"
	case DiagramTypeClass:
		return "classDiagram"
	default:
		return "flowchart TD"
	}
}

func getDiagramPrompt(code string, purpose string, diagramType DiagramType) string {
	switch diagramType {
	case DiagramTypeClass:
		return `다음 코드를 분석하여 상세한 클래스 다이어그램을 Mermaid 형식으로 생성해주세요:

1. 모든 클래스와 그들 간의 상속 관계를 포함할 것
2. 각 클래스에 대해 다음 정보를 명확히 표시할 것:
   - 모든 속성(프로퍼티)과 접근 제한자(public, protected, private)
   - 중요 메서드와 반환 타입
   - 생성자와 매개변수
3. 클래스 간의 관계를 다음과 같이 표시할 것:
   - 상속(inheritance): <|--
   - 합성(composition): *--
   - 집합(aggregation): o--
   - 연관(association): -->
   - 의존성(dependency): ..>
4. 추상 클래스와 인터페이스는 특별히 표시할 것
5. 중요 메서드에는 매개변수 타입과 반환 타입도 포함할 것
6. 패키지 또는 모듈 구조를 나타내는 경계선 추가할 것

다이어그램은 왼쪽에서 오른쪽으로 읽기 쉽게 배치하고, 관련 클래스는 서로 가깝게 배치해주세요.`
	case DiagramTypeSequence:
		return `다음 코드를 분석하여 상세한 시퀀스 다이어그램을 Mermaid 형식으로 생성해주세요:

1. 주요 실행 흐름을 시각화하되, 다음 정보를 포함할 것:
   - 모든 관련 객체/액터/서비스를 참가자로 포함
   - 참가자 간의 모든 메서드 호출과 함수 호출
   - 각 호출에 대한 매개변수 값
   - 반환 값과 데이터 흐름
   - 비동기 호출과 콜백은 점선 화살표로 표시

2. 다음과 같은 고급 기능을 포함할 것:
   - 활성화 상자(activation box)로 메서드 실행 기간 표시
   - 조건부 로직과 분기 처리(alt/opt/loop)
   - 중요 주석과 설명
   - 병렬 처리나 동시성 작업(par)
   - 타임아웃 또는 시간 제한 표시

3. 오류 처리 흐름도 포함할 것:
   - 예외 발생과 처리 경로
   - 롤백 메커니즘
   - 재시도 로직

4. 중요 상태 변경이나 이벤트도 표시할 것

시간순으로 위에서 아래로 흐르는 명확한 다이어그램을 만들고, 복잡한 부분은 단계별로 분리하여 표현해주세요.`
	case DiagramTypeFlowchart:
		return `다음 코드를 분석하여 상세한 플로우차트를 Mermaid 형식으로 생성해주세요:

1. 알고리즘 또는 비즈니스 로직의 전체 흐름을 표현하되, 다음 요소를 명확히 구분할 것:
   - 시작과 종료 지점(둥근 모서리 직사각형)
   - 처리 단계(직사각형)
   - 결정 포인트/조건문(마름모)
   - 입력/출력 작업(평행사변형)
   - 서브루틴 호출(직사각형 + 세로선)
   - 데이터 저장(실린더)

2. 다음과 같은 로직 흐름을 상세히 표현할 것:
   - 모든 조건문(if-else, switch)과 분기 경로
   - 모든 반복문(for, while, do-while)의 시작, 반복 조건, 종료
   - 예외 처리 경로와 에러 핸들링
   - 중요 변수 값 변경 및 상태 업데이트
   - 외부 시스템 호출 및 API 요청

3. 각 노드에 충분한 컨텍스트 정보 포함할 것:
   - 조건문의 정확한 조건식
   - 처리 단계에서 실행되는 주요 로직
   - 변수 할당이나 계산식
   - 함수 호출의 주요 매개변수

4. 복잡한 알고리즘의 경우 다음 내용 추가:
   - 주요 진입/종료 지점 강조
   - 중요 결정 포인트에 주석 추가
   - 성능에 영향을 미치는 부분 식별
   - 재귀 호출이나 순환 참조 명확히 표시

다이어그램은 위에서 아래로, 왼쪽에서 오른쪽으로 논리적 흐름을 따르도록 배치하고, 복잡한 흐름은 하위 다이어그램으로 분리해주세요.`
	}
	return ""
}
