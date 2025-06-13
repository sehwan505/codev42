package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"codev42-agent/util"

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

type DiagramType = util.DiagramType

const (
	DiagramTypeFlowchart = util.DiagramTypeFlowchart
	DiagramTypeSequence  = util.DiagramTypeSequence
	DiagramTypeClass     = util.DiagramTypeClass
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

		validator := util.NewDiagramValidator()
		// 다이어그램 검증
		if err := validator.ValidateDiagram(result.Diagram, diagramType); err != nil {
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
	prompt := fmt.Sprintf(`
다음 규칙을 엄격히 따라주세요:
1. 코드 구조를 명확하게 시각화하는 %s 다이어그램을 생성해주세요
2. 한국어로 내용을 작성해주세요
3. '%s'로 시작하는 Mermaid 문법을 사용해주세요
4. 컴포넌트/함수 간의 관계를 적절하게 보여주세요
5. 다이어그램이 문법적으로 올바르고 의미있는지 확인해주세요 %s

**중요:** 기본적으로는 따옴표 없이 작성해주세요.
mermaid 다이어그램 코드만을 반환하고, 설명이나 추가 텍스트는 포함하지 마세요.`, diagramType, getMermaidPrefix(diagramType), retryNote)

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
		Model:       openai.F(openai.ChatModelGPT4o2024_11_20),
		Temperature: openai.F(0.0),
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
	// 공통 한국어 처리 규칙
	commonKoreanRules := `
**한국어 처리 최적화 규칙 (필수):**
- 따옴표 없이 작성: class 문서처리시스템
- 언더스코어 사용: class 사용자_인터페이스_관리자
- 복합어는 적절히 분리하여 가독성 향상: 데이터베이스접속관리자 → 데이터베이스_접속_관리자`

	switch diagramType {
	case DiagramTypeClass:
		return fmt.Sprintf(`다음 코드를 분석하여 상세한 클래스 다이어그램을 Mermaid 형식으로 생성해주세요.

%s

**클래스 다이어그램 구체적 요구사항:**
1. 클래스 정의 및 관계:
   - 모든 클래스: class 클래스명 (언더스코어 활용)
   - 상속관계: 부모클래스 <|-- 자식클래스
   - 컴포지션: 전체클래스 *-- 부분클래스
   - 집합관계: 컨테이너클래스 o-- 요소클래스
   - 연관관계: 클래스A --> 클래스B : 관계명
   - 의존관계: 클래스A ..> 클래스B : 사용

2. 클래스 내부 구조 표현:
   - 접근제한자: +공개메소드(), -비공개속성, #보호메소드(), ~패키지속성
   - 속성정의: +속성명 타입
   - 메소드정의: +메소드명(매개변수타입) 반환타입
   - 생성자: +생성자명(매개변수목록)
   - 정적멤버: +정적메소드()$ 또는 +정적속성$

3. 고급 표현 요소:
   - 추상클래스: class 추상클래스명 followed by 추상클래스명 : <<abstract>>
   - 인터페이스: class 인터페이스명 followed by 인터페이스명 : <<interface>>
   - 열거형: class 열거형명 followed by 열거형명 : <<enumeration>>

4. 구조화 및 배치:
   - 관련 클래스들을 그룹핑
   - 상속 계층은 위에서 아래로 배치
   - 의존성 화살표 방향 일관성 유지
5. classDiagram 문법을 사용해주세요
   - 항상 classDiagram으로 시작해주세요
   - 한 개 이상의 class를 포함해주세요

코드: %s
목적: %s

mermaid classDiagram 코드만 반환하세요.`, commonKoreanRules, code, purpose)

	case DiagramTypeSequence:
		return fmt.Sprintf(`다음 코드를 분석하여 상세한 시퀀스 다이어그램을 Mermaid 형식으로 생성해주세요:

%s

**시퀀스 다이어그램 구체적 요구사항:**
1. 참가자 정의:
   - participant 참가자명 (언더스코어 활용)
   - participant 시스템_관리자
   - participant 데이터베이스_서버

2. 메시지 타입별 표현:
   - 동기호출: 참가자A->>참가자B: 메시지내용
   - 비동기호출: 참가자A-)+참가자B: 비동기메시지
   - 응답메시지: 참가자B-->>참가자A: 응답내용
   - 자기호출: 참가자A->>참가자A: 내부처리

3. 실행 제어 구조:
   - 활성화: activate 참가자명, deactivate 참가자명
   - 조건문: alt 조건1, else 조건2, end
   - 반복문: loop 반복조건, end
   - 선택문: opt 선택조건, end
   - 병렬처리: par 병렬작업1, and 병렬작업2, end

4. 주석 및 설명:
   - Note over 참가자: 설명내용
   - Note left of 참가자: 왼쪽설명
   - Note right of 참가자: 오른쪽설명

5. 에러 처리:
   - 예외발생과 처리경로 명시
   - 타임아웃 상황 표현
   - 재시도 로직 포함

코드: %s
목적: %s

mermaid sequenceDiagram 코드만 반환하세요.`, commonKoreanRules, code, purpose)

	case DiagramTypeFlowchart:
		return fmt.Sprintf(`다음 코드를 분석하여 상세한 플로우차트를 Mermaid 형식으로 생성해주세요:

%s

**플로우차트 구체적 요구사항:**
1. 노드 타입별 표현:
   - 프로세스: A[처리과정명]
   - 결정점: B{조건확인}
   - 시작/종료: C((시작점)), D((종료점))
   - 입출력: E[/데이터입력/], F[\데이터출력\]
   - 서브루틴: G[[서브루틴호출]]
   - 데이터저장: H[(데이터베이스)]

2. 연결 및 흐름:
   - 기본연결: A --> B
   - 조건부연결: B -->|조건참| C
   - 라벨연결: C -->|처리과정| D
   - 점선연결: E -.-> F

3. 논리 구조 표현:
   - 모든 if-else 분기 명시
   - 반복문의 시작조건과 종료조건
   - switch문의 모든 case 분기
   - 예외처리 경로 포함

4. 서브그래프 활용:
   - subgraph 영역명
   - 관련 노드들 그룹핑
   - end로 서브그래프 종료
   - 예: subgraph 인증_처리_영역

5. 고급 표현:
   - 병렬처리 경로 표시
   - 재귀호출 표현
   - 상태변경 지점 강조
   - 성능 크리티컬 구간 식별

코드: %s
목적: %s

mermaid flowchart TD 코드만 반환하세요.`, commonKoreanRules, code, purpose)
	}
	return ""
}
