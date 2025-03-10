package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

type Annotation struct {
	Name        string `json:"name" jsonschema_description:"The name of the function or method"`
	Params      string `json:"params" jsonschema_description:"The parameters of the function with types"`
	Returns     string `json:"returns" jsonschema_description:"The return value of the function with type"`
	Description string `json:"description" jsonschema_description:"The description of the function"`
}

type Plan struct {
	ClassName   string       `json:"class_name" jsonschema_description:"class name if empty then it is function"`
	Annotations []Annotation `json:"annotations" jsonschema_description:"Structured annotations for functions and class methods"`
}

type DevPlan struct {
	Language string `json:"language" jsonschema_description:"The programming language for development"`
	Plans    []Plan `json:"plans" jsonschema_description:"List of development plans with class Name and annotations"`
}

type MasterAgent struct {
	Client *openai.Client
}

func NewMasterAgent(apiKey string) *MasterAgent {
	openaiClient := GetClient(apiKey)

	return &MasterAgent{
		Client: openaiClient.Client(),
	}
}

func GenerateDevPlanSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var DevPlanResponseSchema = GenerateDevPlanSchema[DevPlan]()

func (agent MasterAgent) Call(prompt string) (*DevPlan, error) {
	prompt = "prompt: " + prompt
	prompt += `
	you should follow the rules to make a development plan
	rule: make a development plan about prompt with annotations of functions and classes as a list
	annotation follow @name, @params, @returns, @description, 
	if the development is for a class, ClassName should be given, then annotations should be list of methods
	or if the development is for a function, ClassName should be empty, then annotations should be list with only one item
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
