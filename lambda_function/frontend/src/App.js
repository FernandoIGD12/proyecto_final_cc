import React, { useState } from 'react';
import './App.css';

function App() {
  const [formData, setFormData] = useState({
    velocidad_sag_rpm: '',
    flujo_de_agua_m3_h: '',
    ugm1: '',
    ugm2: '',
    ugm3: '',
    porc_grueso: '',
    porc_intermedio: '',
    porc_fino: '',
  });
  const [prediction, setPrediction] = useState(null);
  const [error, setError] = useState(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setPrediction(null);

    // Replace with your actual API Gateway endpoint
    const apiGatewayUrl = 'https://3roqr2e3c0.execute-api.us-east-1.amazonaws.com/default/sagemaker_prediction';

    const parsedData = Object.keys(formData).reduce((acc, key) => {
      acc[key] = parseFloat(formData[key]);
      return acc;
    }, {});

    try {
      const response = await fetch(apiGatewayUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(parsedData),
      });

      if (!response.ok) {
        throw new Error('Something went wrong');
      }

      const result = await response.json();
      setPrediction(result);
    } catch (error) {
      setError(error.message);
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Tonnage per hour Prediction</h1>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            {Object.keys(formData).map((key) => (
              <div key={key} className="form-field">
                <label htmlFor={key}>{key.replace(/_/g, ' ')}</label>
                <input
                  type="number"
                  id={key}
                  name={key}
                  value={formData[key]}
                  onChange={handleChange}
                  required
                />
              </div>
            ))}
          </div>
          <button type="submit">Get Prediction</button>
        </form>
        {prediction && (
          <div className="prediction-result">
            <h2>Prediction Result</h2>
            <p>Predicted Rendimiento: {prediction.predicted_rendimiento_t_h}</p>
            <p>Model Used: {prediction.model_used}</p>
          </div>
        )}
        {error && (
          <div className="error-message">
            <h2>Error</h2>
            <p>{error}</p>
          </div>
        )}
      </header>
    </div>
  );
}

export default App;
