package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/sehwan505/codev42/configs"
)

func WorkerAgent(config *configs.Config, buildPlan string) (string, error) {
	openaiClient := GetClient(config.OpenAiKey).Client()

	chatCompletion, err := openaiClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI completion error: %v", err)
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
