package main

import (
	"fmt"

	"github.com/sehwan505/codev42/configs"
	"github.com/sehwan505/codev42/internal/agent"
)

func main() {
	config, err := configs.GetConfig()
	if err != nil {
		return
	}
	masterAgent := agent.NewMasterAgent(config.OpenAiKey)
	var prompt string
	prompt = "make number baseball with python"
	fmt.Printf("Prompt: %s\n", prompt)
	devPlan, err := masterAgent.Call(prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	workerAgent := agent.NewWorkerAgent(config.OpenAiKey)
	for _, annotation := range devPlan.Annotations {
		fmt.Printf("Function: %s\n", annotation)
		devResult, err := workerAgent.Call(annotation)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Code: %s\n", devResult.Code)
	}
}
