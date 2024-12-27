package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

type DevPlan struct {
	Language    string   `json:"language" jsonschema_description:"The programming language for development"`
	Annotations []string `json:"annotations" jsonschema_description:"annotations of functions and classes for planning development with params, returns, and description"`
}

// type Annotation struct {
// 	params      string `json:"params" jsonschema_description:"The parameters of the function with types"`
// 	returns     string `json:"returns" jsonschema_description:"The return value of the function with type"`
// 	description string `json:"description" jsonschema_description:"The description of the function"`
// }

type MasterAgent struct {
	Client *openai.Client
}

func NewMasterAgent(apiKey string) *MasterAgent {
	openaiClient := GetClient(apiKey)

	return &MasterAgent{
		Client: openaiClient.Client(),
	}
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var DevPlanResponseSchema = GenerateSchema[DevPlan]()

func (agent MasterAgent) Call(prompt string) (*DevPlan, error) {
	prompt = "prompt: " + prompt
	prompt += `
	rule: make a development plan about prompt with annotations of functions and classes as a list
	every annotation must contains @params, @returns, @description	
	`
	print("> ")
	println(prompt)

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("development_plan"),
		Description: openai.F("A development plan with annotations of functions and classes"),
		Schema:      openai.F(DevPlanResponseSchema),
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

	devPlan := &DevPlan{}
	fmt.Printf("Chat: %v\n", chat)
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), devPlan)
	if err != nil {
		return nil, err
	}
	return devPlan, nil
}
