package main

import (
	"fmt"

	"sync"

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
	var wg sync.WaitGroup
	resultChan := make(chan *agent.DevResult, len(devPlan.Annotations))
	errorChan := make(chan error, len(devPlan.Annotations))
	for _, annotation := range devPlan.Annotations {
		wg.Add(1)
		go func(annotation string) {
			defer wg.Done()
			fmt.Printf("Function: %s\n", annotation)
			devResult, err := workerAgent.Call(annotation)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				errorChan <- err
				return
			}
			resultChan <- devResult
		}(annotation)
	}

	wg.Wait()
	close(resultChan)
	close(errorChan)

	var results []*agent.DevResult
	for result := range resultChan {
		results = append(results, result)
	}
	for _, result := range results {
		fmt.Printf("result: %v\n\n", result)
	}
}
