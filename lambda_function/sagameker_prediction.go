package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

// Global SageMaker client for reuse across Lambda invocations
var smRuntimeClient *sagemakerruntime.Client

func init() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// Log the error and allow Lambda to crash, which signals an irrecoverable setup failure
		fmt.Printf("failed to load AWS config: %v\n", err)
		return
	}
	// Create a SageMaker Runtime client
	smRuntimeClient = sagemakerruntime.NewFromConfig(cfg)
}

func main() {
	lambda.Start(handler)
}

// handler is the main function that gets called by the Lambda runtime
func handler(ctx context.Context, request *events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			},
			Body: "",
		}, nil
	}


	// --- Configuration Check ---
	endpointName := os.Getenv("SAGEMAKER_ENDPOINT_NAME")
	if endpointName == "" {
		return errorResponse(500, "SAGEMAKER_ENDPOINT_NAME environment variable not set"), nil
	}
	if smRuntimeClient == nil {
		return errorResponse(500, "SageMaker client failed to initialize"), nil
	}

	// --- 1. Parse Input JSON ---
	var inputData PredictionInput
	if err := json.Unmarshal([]byte(request.Body), &inputData); err != nil {
		fmt.Printf("Error unmarshalling request body: %v\n", err)
		return errorResponse(400, "Invalid JSON input format"), nil
	}

	// --- 2. Format Data for SageMaker (CSV String) ---
	// The order MUST match the features the model was trained with (excluding the target column).
	csvData := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v",
		inputData.VelocidadSAGRpm,
		inputData.FlujoDeAguaM3H,
		inputData.UGM1,
		inputData.UGM2,
		inputData.UGM3,
		inputData.PorcGrueso,
		inputData.PorcIntermedio,
		inputData.PorcFino,
	)
	fmt.Printf("Payload sent to SageMaker: %s\n", csvData)

	// --- 3. Invoke SageMaker Endpoint ---
	// Your endpoint was set up to accept CSV data
	input := &sagemakerruntime.InvokeEndpointInput{
		EndpointName: aws.String(endpointName),
		ContentType:  aws.String("text/csv"),
		Body:         []byte(csvData),
		// We expect CSV back because we used the CSV Deserializer in the predictor testing.
		// If you used JSON Deserializer, change this to "application/json".
		Accept: aws.String("text/csv"),
	}

	result, err := smRuntimeClient.InvokeEndpoint(ctx, input)
	if err != nil {
		fmt.Printf("SageMaker InvokeEndpoint error: %v\n", err)
		return errorResponse(500, fmt.Sprintf("Error calling SageMaker: %v", err.Error())), nil
	}

	// --- 4. Process Output ---
	// The response body is the raw CSV prediction, e.g., "1221.59"
	predictionStr := strings.TrimSpace(string(result.Body))

	prediction, err := strconv.ParseFloat(predictionStr, 64)
	if err != nil {
		fmt.Printf("Error parsing prediction result '%s': %v\n", predictionStr, err)
		return errorResponse(500, "Could not parse prediction result from model"), nil
	}

	// --- 5. Return Success Response ---
	responseBody, _ := json.Marshal(PredictionOutput{
		PredictedRendimiento: prediction,
		ModelUsed:            endpointName,
	})

	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{
			"Content-Type": "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		},
		Body:            string(responseBody),
		IsBase64Encoded: false,
	}, nil
}

// Helper function to generate a standard error response
func errorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		Headers:         map[string]string{
			"Content-Type": "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		},
		Body:            string(body),
		IsBase64Encoded: false,
	}
}
