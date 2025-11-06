package main

// PredictionInput struct defines the expected JSON structure from API Gateway/Postman.
// The fields must match the order of features used by your SageMaker model (Velocidad_SAG_rpm, Flujo_de_agua_m3_h, UGM1, etc.)
type PredictionInput struct {
	VelocidadSAGRpm float64 `json:"velocidad_sag_rpm"`
	FlujoDeAguaM3H  float64 `json:"flujo_de_agua_m3_h"`
	UGM1            float64 `json:"ugm1"`
	UGM2            float64 `json:"ugm2"`
	UGM3            float64 `json:"ugm3"`
	PorcGrueso      float64 `json:"porc_grueso"`
	PorcIntermedio  float64 `json:"porc_intermedio"`
	PorcFino        float64 `json:"porc_fino"`
}

// PredictionOutput defines the structure for the JSON response back to the client.
type PredictionOutput struct {
	PredictedRendimiento float64 `json:"predicted_rendimiento_t_h"`
	ModelUsed            string  `json:"model_used"`
}
