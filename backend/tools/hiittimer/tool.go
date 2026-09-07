package hiittimer

import (
	"fmt"
	"math"

	"developer-toolbox/backend/models"
)

type Tool struct{}

func NewTool() *Tool { return &Tool{} }

func (t *Tool) Info() models.Tool {
	return models.Tool{
		ID:           "hiit-timer",
		Name:         "HIIT Timer",
		Description:  "Configurable work/rest interval timer with rounds and a breathing-light display.",
		Category:     "utility",
		Icon:         "timer",
		Version:      "1.0.0",
		Capabilities: models.ToolCapabilities{Stateful: true},
		Keywords:     []string{"hiit", "timer", "interval", "workout", "tabata", "training", "rest"},
	}
}

func (t *Tool) Execute(input models.ToolInput) models.ToolOutput {
	work, err := boundedInt(input.Payload["workSeconds"], 30, 1, 3600)
	if err != nil {
		return models.Failure("INVALID_HIIT_CONFIG", "workSeconds must be an integer from 1 to 3600")
	}
	rest, err := boundedInt(input.Payload["restSeconds"], 15, 1, 3600)
	if err != nil {
		return models.Failure("INVALID_HIIT_CONFIG", "restSeconds must be an integer from 1 to 3600")
	}
	rounds, err := boundedInt(input.Payload["rounds"], 8, 1, 99)
	if err != nil {
		return models.Failure("INVALID_HIIT_CONFIG", "rounds must be an integer from 1 to 99")
	}
	return models.ToolOutput{Success: true, Data: map[string]any{
		"workSeconds":  work,
		"restSeconds":  rest,
		"rounds":       rounds,
		"intervals":    rounds*2 - 1,
		"totalSeconds": work*rounds + rest*(rounds-1),
	}}
}

func boundedInt(value any, fallback, min, max int) (int, error) {
	if value == nil {
		return fallback, nil
	}
	var result int
	switch number := value.(type) {
	case int:
		result = number
	case float64:
		if math.Trunc(number) != number {
			return 0, fmt.Errorf("not an integer")
		}
		result = int(number)
	default:
		return 0, fmt.Errorf("not a number")
	}
	if result < min || result > max {
		return 0, fmt.Errorf("out of range")
	}
	return result, nil
}
