package appgitlab

var globalStageOrder = []string{"init", "build", "test", "lint", "scan", "package", "publish", "deploy", "cleanup"}

func getOrderedStages(allStages []string) []string {
	seen := make(map[string]bool)
	for _, s := range allStages {
		seen[s] = true
	}

	var orderedStages []string
	for _, stage := range globalStageOrder {
		if seen[stage] {
			orderedStages = append(orderedStages, stage)
		}
	}

	return orderedStages
}
