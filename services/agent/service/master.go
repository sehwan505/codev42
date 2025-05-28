package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

type Annotation struct {
	Name        string `json:"name" jsonschema_description:"함수나 메서드의 이름"`
	Params      string `json:"params" jsonschema_description:"타입이 포함된 함수의 매개변수들"`
	Returns     string `json:"returns" jsonschema_description:"타입이 포함된 함수의 반환값"`
	Description string `json:"description" jsonschema_description:"함수에 대한 설명"`
}

type Plan struct {
	ClassName   string       `json:"class_name" jsonschema_description:"클래스 이름 (비어있으면 함수)"`
	Annotations []Annotation `json:"annotations" jsonschema_description:"함수와 클래스 메서드에 대한 구조화된 어노테이션"`
}

type DevPlan struct {
	Language string `json:"language" jsonschema_description:"개발에 사용될 프로그래밍 언어"`
	Plans    []Plan `json:"plans" jsonschema_description:"클래스 이름과 어노테이션이 포함된 개발 계획 목록"`
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
	prompt = "프롬프트: " + prompt
	prompt += `
	다음 규칙에 따라 개발 계획을 수립해야 합니다
	규칙: 프롬프트에 대해 함수와 클래스의 어노테이션을 포함한 개발 계획을 목록으로 작성하세요
	어노테이션은 @name, @params, @returns, @description을 따릅니다
	클래스를 위한 개발인 경우, ClassName을 제공하고 어노테이션은 메소드 목록이어야 합니다
	함수를 위한 개발인 경우, ClassName은 비워두고 어노테이션은 하나의 항목만 포함하는 목록이어야 합니다
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
	// DevPlan을 model.DevPlan으로 변환
	return devPlan, nil
}
