package actionsdk

type HealthV1Response struct {
	Status string `json:"status"`
}

type LogV1Request struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context"`
}
