package relay

type ModelPricing struct {
	InputPrice  int64
	OutputPrice int64
	PriceUnit   int
}

func CalculateCost(pricing *ModelPricing, inputTokens, outputTokens int) int64 {
	if pricing == nil || pricing.PriceUnit == 0 {
		return 0
	}

	inputCost := int64(inputTokens) * pricing.InputPrice / int64(pricing.PriceUnit)
	outputCost := int64(outputTokens) * pricing.OutputPrice / int64(pricing.PriceUnit)

	return inputCost + outputCost
}

func GetDefaultPricing(model string) *ModelPricing {
	prices := map[string]*ModelPricing{
		"gpt-4o":            {InputPrice: 2500, OutputPrice: 10000, PriceUnit: 1000000},
		"gpt-4o-mini":       {InputPrice: 150, OutputPrice: 600, PriceUnit: 1000000},
		"claude-sonnet-4-6": {InputPrice: 3000, OutputPrice: 15000, PriceUnit: 1000000},
		"claude-haiku-4-5":  {InputPrice: 800, OutputPrice: 4000, PriceUnit: 1000000},
		"gemini-2.5-pro":    {InputPrice: 1250, OutputPrice: 10000, PriceUnit: 1000000},
		"gemini-2.5-flash":  {InputPrice: 150, OutputPrice: 600, PriceUnit: 1000000},
		"deepseek-chat":     {InputPrice: 270, OutputPrice: 1100, PriceUnit: 1000000},
		"deepseek-reasoner": {InputPrice: 550, OutputPrice: 2190, PriceUnit: 1000000},
	}

	if p, ok := prices[model]; ok {
		return p
	}
	return &ModelPricing{InputPrice: 1000, OutputPrice: 3000, PriceUnit: 1000000}
}
