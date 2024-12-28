package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

type DevResult struct {
	Description string   `json:"description" jsonschema_description:"description of the development result"`
	Code        []string `json:"annotations" jsonschema_description:"code implemented from the dev plan"`
}

// type Annotation struct {
// 	params      string `json:"params" jsonschema_description:"The parameters of the function with types"`
// 	returns     string `json:"returns" jsonschema_description:"The return value of the function with type"`
// 	description string `json:"description" jsonschema_description:"The description of the function"`
// }

type WorkerAgent struct {
	Client *openai.Client
}

func NewWorkerAgent(apiKey string) *WorkerAgent {
	openaiClient := GetClient(apiKey)

	return &WorkerAgent{
		Client: openaiClient.Client(),
	}
}

func GenerateDevResultSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var DevResultResponseSchema = GenerateDevResultSchema[DevResult]()

func (agent WorkerAgent) Call(devPlan string) (*DevResult, error) {
	prompt := "dev plan: " + devPlan
	prompt += `
	you should follow the dev plan to make a development result
	follow the dev plan to make a development result
	development result must contains the code and description of the development result by following the dev plan
	`
	print("> ")
	println(prompt)

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("development_result"),
		Description: openai.F("code and description of development result from the dev plan"),
		Schema:      openai.F(DevResultResponseSchema),
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

	devResult := &DevResult{}
	fmt.Printf("Chat: %v\n", chat)
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), devResult)
	if err != nil {
		return nil, err
	}
	return devResult, nil
}
