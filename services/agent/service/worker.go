package service

import (
	"codev42-agent/model"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

type ImplementResult struct {
	Code string `json:"code" jsonschema_description:"code implemented from the dev plan"`
}

type WorkerAgent struct {
	Client *openai.Client
}

func NewWorkerAgent(apiKey string) *WorkerAgent {
	openaiClient := GetClient(apiKey)

	return &WorkerAgent{
		Client: openaiClient.Client(),
	}
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

func (agent WorkerAgent) call(language string, devPlan string) (*ImplementResult, error) {
	prompt := "dev plan: " + devPlan
	prompt += "language: " + language
	prompt += `
	you should follow the dev plan to make a development result
	follow the dev plan to make a development result
	development result must contains the code and description of the development result by following the dev plan
	`
	print("> ")
	println(prompt)

	var ImplementResultResponseSchema = GenerateImplementResultSchema[ImplementResult]()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("development_result"),
		Description: openai.F("code and description of development result from the dev plan"),
		Schema:      openai.F(ImplementResultResponseSchema),
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
		Model: openai.F(openai.ChatModelGPT4oMini),
	})

	if err != nil {
		return nil, err
	}

	ImplementResult := &ImplementResult{}
	fmt.Printf("Chat: %v\n", chat)
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), ImplementResult)
	if err != nil {
		return nil, err
	}
	return ImplementResult, nil
}

func (agent WorkerAgent) ImplementPlan(language string, plans []model.Plan) ([]*ImplementResult, error) {
	var wg sync.WaitGroup
	resultChan := make(chan *ImplementResult, len(plans))
	errorChan := make(chan error, len(plans))

	for _, plan := range plans {
		wg.Add(1)
		go func(plan model.Plan) {
			defer wg.Done()
			fmt.Printf("Processing: %s\n", plan.Annotations)
			planString := "className: " + plan.ClassName + "\n"
			for _, annotation := range plan.Annotations {
				planString += "functionName: " + annotation.Name + "\n"
				planString += "functionDescription: " + annotation.Description + "\n"
				planString += "functionParameters: " + annotation.Params + "\n"
				planString += "functionReturnType: " + annotation.Returns + "\n"
			}
			ImplementResult, err := agent.call(language, planString)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- ImplementResult
		}(plan)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	var results []*ImplementResult
	for result := range resultChan {
		results = append(results, result)
	}
	if len(errorChan) > 0 {
		var errors []string
		for err := range errorChan {
			errors = append(errors, err.Error())
		}
		return nil, fmt.Errorf("failed to implement plan: %v", errors)
	}
	return results, nil
}
