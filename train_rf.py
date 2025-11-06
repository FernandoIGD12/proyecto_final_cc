import argparse
import os
import pandas as pd
from sklearn.ensemble import RandomForestRegressor
import joblib
import json
import numpy as np
import io


def model_fn(model_dir):
    """
    Loads the saved model from the model directory.
    """
    print("Loading model...")
    model = joblib.load(os.path.join(model_dir, "model.joblib"))
    print("Model loaded.")
    return model


def input_fn(request_body, request_content_type):
    """
    Deserializes the input data. We'll support JSON and CSV.
    """
    print(f"Received request with type: {request_content_type}")
    if request_content_type == "application/json":
        data = json.loads(request_body)
        # Assuming JSON is a list or list of lists
        return np.array(data)
    elif request_content_type == "text/csv":
        # Read CSV data into a numpy array
        return np.loadtxt(io.StringIO(request_body), delimiter=',')
    elif request_content_type == "application/x-npy":
        return np.load(io.BytesIO(request_body))
    else:
        raise ValueError(f"Unsupported content type: {request_content_type}")


def predict_fn(input_data, model):
    """
    Makes a prediction using the loaded model.
    """
    print(f"Making prediction on data of shape: {input_data.shape}")
    prediction = model.predict(input_data)
    return prediction


def output_fn(prediction, accept_type):
    """
    Serializes the prediction output.
    """
    print(f"Serializing prediction for accept type: {accept_type}")
    if accept_type == "application/json" or accept_type == "application/json; charset=utf-8":
        # Convert numpy array to list for JSON serialization
        return json.dumps(prediction.tolist()), "application/json"
    elif accept_type == "text/csv":
        return '\n'.join(map(str, prediction)), "text/csv"
    else:
        raise ValueError(f"Unsupported accept type: {accept_type}")


# Training function
if __name__ == "__main__":
    
    parser = argparse.ArgumentParser()
    parser.add_argument("--n-estimators", type=int, default=400)
    parser.add_argument("--random-state", type=int, default=42)
    parser.add_argument("--model-dir", type=str, default=os.environ.get("SM_MODEL_DIR"))
    parser.add_argument("--train", type=str, default=os.environ.get("SM_CHANNEL_TRAIN"))
    
    args, _ = parser.parse_known_args()

    print("--- Starting Training ---")
    train_file_path = os.path.join(args.train, "train.csv")
    df = pd.read_csv(train_file_path)
    
    target_col = "rendimiento_t_h"
    X_train = df.drop(target_col, axis=1)
    y_train = df[target_col]
    
    print(f"Loaded training data. Shape: {X_train.shape}")

    rf = RandomForestRegressor(
        n_estimators=args.n_estimators,
        random_state=args.random_state,
        n_jobs=-1
    )
    
    print("Training Random Forest model...")
    rf.fit(X_train, y_train)
    print("Training complete.")

    joblib.dump(rf, os.path.join(args.model_dir, "model.joblib"))
    print(f"Model saved to {args.model_dir}/model.joblib")
    print("--- Training Finished ---")