# Parallel Notebooks Example

This example demonstrates how to use notebook-use to run multiple Jupyter notebooks in parallel.

## Example Usage

### 1. Create Multiple Notebook Environments

```python
# Create three notebook environments for different experiments
notebooks = [
    {"name": "model-training-v1", "path": "train_model.ipynb", "kernel_spec": "python3"},
    {"name": "model-training-v2", "path": "train_model.ipynb", "kernel_spec": "python3"},
    {"name": "data-analysis", "path": "analyze_data.ipynb", "kernel_spec": "python3"}
]
```

### 2. Run Notebooks in Parallel

Using the MCP tool `notebook_parallel_run`:

```json
{
    "notebooks": [
        {"name": "model-training-v1", "path": "experiments/train_model.ipynb"},
        {"name": "model-training-v2", "path": "experiments/train_model.ipynb"},
        {"name": "data-analysis", "path": "experiments/analyze_data.ipynb"}
    ],
    "max_parallel": 3
}
```

### 3. Benefits

- **Parallel Experimentation**: Run multiple experiments with different hyperparameters simultaneously
- **A/B Testing**: Compare different approaches side by side
- **Resource Efficiency**: Maximize hardware utilization
- **State Isolation**: Each notebook maintains its own kernel state
- **Git Integration**: All changes are tracked in separate git branches

### 4. Example Notebook Content

```python
# Cell 1: Import and Setup
import numpy as np
import time
from datetime import datetime

experiment_id = datetime.now().strftime("%Y%m%d_%H%M%S")
print(f"Starting experiment: {experiment_id}")

# Cell 2: Run Computation
def train_model(data_size=1000, epochs=10):
    # Simulate model training
    data = np.random.randn(data_size, 10)
    weights = np.random.randn(10)
    
    for epoch in range(epochs):
        # Simulate training step
        loss = np.mean((data @ weights) ** 2)
        weights -= 0.01 * (data.T @ (data @ weights))
        print(f"Epoch {epoch}: Loss = {loss:.4f}")
    
    return weights, loss

# Cell 3: Execute and Save Results
start_time = time.time()
final_weights, final_loss = train_model(data_size=5000, epochs=20)
training_time = time.time() - start_time

results = {
    'experiment_id': experiment_id,
    'final_loss': float(final_loss),
    'training_time': training_time,
    'weights_norm': float(np.linalg.norm(final_weights))
}

print(f"\nTraining completed in {training_time:.2f} seconds")
print(f"Final loss: {final_loss:.6f}")
```

### 5. Aggregating Results

After parallel execution, results from all notebooks can be aggregated:

```python
# Collect results from all parallel runs
all_results = []
for notebook in executed_notebooks:
    result = notebook.get_state()
    all_results.append({
        'name': notebook.name,
        'id': notebook.id,
        'results': result.variables.get('results', {}),
        'execution_time': result.last_executed
    })

# Compare results
best_model = min(all_results, key=lambda x: x['results'].get('final_loss', float('inf')))
print(f"Best model: {best_model['name']} with loss: {best_model['results']['final_loss']}")
```

## Architecture Benefits

1. **Scalability**: Can run hundreds of notebooks simultaneously
2. **Reproducibility**: Each notebook execution is tracked in git
3. **Isolation**: Notebooks run in separate containers with isolated state
4. **Flexibility**: Support for different kernels (Python, R, Julia)
5. **Integration**: MCP tools enable AI assistants to manage notebook execution