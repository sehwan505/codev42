package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

type CombinedResult struct {
	Code string `json:"code" jsonschema_description:"the result of combining the codes"`
}

type AnalyserAgent struct {
	Client *openai.Client
}

func NewAnalyserAgent(apiKey string) *AnalyserAgent {
	openaiClient := GetClient(apiKey)

	return &AnalyserAgent{
		Client: openaiClient.Client(),
	}
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

type DiagramResult struct {
	Diagram string `json:"flow chart diagram"`
}

func (agent AnalyserAgent) ImplementDiagram(code string) (string, error) {
	prompt := "Please analyze the following code and create a Mermaid flowchart diagram showing function relationships. Code:\n" + code
	prompt += `
	Please follow these rules:
	1. Create a flowchart showing all functions and their relationships/dependencies
	2. Include a brief description of each function's purpose
	3. Show the flow of data/control between functions using arrows
	4. Use Mermaid syntax to generate the flowchart, starting with 'flowchart TD'
	5. Add notes or subgraphs to group related functions
	6. The output should ONLY contain the Mermaid diagram code
	7. Do not include any explanatory text or markdown formatting
	
	Return the result in JSON format with the following structure:
	{
		"diagram": "your mermaid diagram code here"
	}
	`

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("diagram_result"),
		Description: openai.F("mermaid diagram code"),
		Schema:      openai.F(GenerateImplementResultSchema[DiagramResult]()),
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
		return "", err
	}

	var diagramResult DiagramResult
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &diagramResult)
	if err != nil {
		return "", fmt.Errorf("failed to implement diagram: %v", err)
	}

	return diagramResult.Diagram, nil
}
