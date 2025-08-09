package testdata

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// NotebookFixture represents a test notebook
type NotebookFixture struct {
	Cells    []CellFixture          `json:"cells"`
	Metadata map[string]interface{} `json:"metadata"`
	NBFormat int                    `json:"nbformat"`
	NBFormatMinor int               `json:"nbformat_minor"`
}

// CellFixture represents a test cell
type CellFixture struct {
	CellType   string   `json:"cell_type"`
	Source     []string `json:"source"`
	Outputs    []Output `json:"outputs,omitempty"`
	ExecutionCount *int `json:"execution_count,omitempty"`
}

// Output represents cell output
type Output struct {
	OutputType string                 `json:"output_type"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Text       []string               `json:"text,omitempty"`
	Name       string                 `json:"name,omitempty"`
}

// CreateSimpleNotebook creates a basic test notebook
func CreateSimpleNotebook() *NotebookFixture {
	return &NotebookFixture{
		Cells: []CellFixture{
			{
				CellType: "markdown",
				Source:   []string{"# Test Notebook\n", "This is a test notebook for unit testing.\n"},
			},
			{
				CellType: "code",
				Source:   []string{"# Initialize\n", "x = 42\n", "print(f'x = {x}')\n"},
			},
			{
				CellType: "code",
				Source:   []string{"# Compute\n", "y = x * 2\n", "print(f'y = {y}')\n"},
			},
		},
		Metadata: map[string]interface{}{
			"kernelspec": map[string]interface{}{
				"display_name": "Python 3",
				"language":     "python",
				"name":         "python3",
			},
			"language_info": map[string]interface{}{
				"name":    "python",
				"version": "3.9.0",
			},
		},
		NBFormat:      4,
		NBFormatMinor: 5,
	}
}

// CreateDataScienceNotebook creates a data science test notebook
func CreateDataScienceNotebook() *NotebookFixture {
	return &NotebookFixture{
		Cells: []CellFixture{
			{
				CellType: "markdown",
				Source:   []string{"# Data Science Example\n"},
			},
			{
				CellType: "code",
				Source: []string{
					"import numpy as np\n",
					"import pandas as pd\n",
					"import matplotlib.pyplot as plt\n",
				},
			},
			{
				CellType: "code",
				Source: []string{
					"# Generate sample data\n",
					"np.random.seed(42)\n",
					"data = np.random.randn(100, 2)\n",
					"df = pd.DataFrame(data, columns=['x', 'y'])\n",
					"print(df.head())\n",
				},
			},
			{
				CellType: "code",
				Source: []string{
					"# Visualize data\n",
					"plt.figure(figsize=(8, 6))\n",
					"plt.scatter(df['x'], df['y'], alpha=0.5)\n",
					"plt.xlabel('X values')\n",
					"plt.ylabel('Y values')\n",
					"plt.title('Random Data Scatter Plot')\n",
					"plt.show()\n",
				},
			},
			{
				CellType: "code",
				Source: []string{
					"# Compute statistics\n",
					"stats = {\n",
					"    'mean_x': df['x'].mean(),\n",
					"    'mean_y': df['y'].mean(),\n",
					"    'std_x': df['x'].std(),\n",
					"    'std_y': df['y'].std(),\n",
					"    'correlation': df['x'].corr(df['y'])\n",
					"}\n",
					"print('Statistics:', stats)\n",
				},
			},
		},
		Metadata: map[string]interface{}{
			"kernelspec": map[string]interface{}{
				"display_name": "Python 3",
				"language":     "python",
				"name":         "python3",
			},
		},
		NBFormat:      4,
		NBFormatMinor: 5,
	}
}

// CreateMLTrainingNotebook creates a machine learning training notebook
func CreateMLTrainingNotebook() *NotebookFixture {
	return &NotebookFixture{
		Cells: []CellFixture{
			{
				CellType: "markdown",
				Source:   []string{"# ML Model Training\n"},
			},
			{
				CellType: "code",
				Source: []string{
					"import numpy as np\n",
					"from sklearn.model_selection import train_test_split\n",
					"from sklearn.linear_model import LogisticRegression\n",
					"from sklearn.metrics import accuracy_score\n",
				},
			},
			{
				CellType: "code",
				Source: []string{
					"# Generate synthetic dataset\n",
					"np.random.seed(42)\n",
					"X = np.random.randn(1000, 10)\n",
					"y = (X[:, 0] + X[:, 1] - X[:, 2] + np.random.randn(1000) * 0.1 > 0).astype(int)\n",
					"print(f'Dataset shape: {X.shape}')\n",
					"print(f'Class distribution: {np.bincount(y)}')\n",
				},
			},
			{
				CellType: "code",
				Source: []string{
					"# Split data\n",
					"X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)\n",
					"print(f'Training set: {X_train.shape}')\n",
					"print(f'Test set: {X_test.shape}')\n",
				},
			},
			{
				CellType: "code",
				Source: []string{
					"# Train model\n",
					"model = LogisticRegression(random_state=42)\n",
					"model.fit(X_train, y_train)\n",
					"\n",
					"# Evaluate\n",
					"train_acc = accuracy_score(y_train, model.predict(X_train))\n",
					"test_acc = accuracy_score(y_test, model.predict(X_test))\n",
					"\n",
					"print(f'Training accuracy: {train_acc:.3f}')\n",
					"print(f'Test accuracy: {test_acc:.3f}')\n",
				},
			},
		},
		Metadata: map[string]interface{}{
			"kernelspec": map[string]interface{}{
				"display_name": "Python 3",
				"language":     "python",
				"name":         "python3",
			},
		},
		NBFormat:      4,
		NBFormatMinor: 5,
	}
}

// CreateErrorNotebook creates a notebook with intentional errors
func CreateErrorNotebook() *NotebookFixture {
	return &NotebookFixture{
		Cells: []CellFixture{
			{
				CellType: "code",
				Source:   []string{"# This will work\n", "x = 10\n"},
			},
			{
				CellType: "code",
				Source:   []string{"# This will cause an error\n", "print(undefined_variable)\n"},
			},
			{
				CellType: "code",
				Source:   []string{"# This will cause a different error\n", "1 / 0\n"},
			},
		},
		Metadata: map[string]interface{}{
			"kernelspec": map[string]interface{}{
				"display_name": "Python 3",
				"language":     "python",
				"name":         "python3",
			},
		},
		NBFormat:      4,
		NBFormatMinor: 5,
	}
}

// SaveNotebook saves a notebook fixture to a file
func SaveNotebook(notebook *NotebookFixture, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(notebook, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// LoadNotebook loads a notebook from a file
func LoadNotebook(path string) (*NotebookFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var notebook NotebookFixture
	err = json.Unmarshal(data, &notebook)
	if err != nil {
		return nil, err
	}
	
	return &notebook, nil
}

// CreateTestWorkspace creates a temporary workspace with test notebooks
func CreateTestWorkspace(dir string) error {
	notebooks := map[string]*NotebookFixture{
		"simple.ipynb":       CreateSimpleNotebook(),
		"data_science.ipynb": CreateDataScienceNotebook(),
		"ml_training.ipynb":  CreateMLTrainingNotebook(),
		"errors.ipynb":       CreateErrorNotebook(),
	}
	
	for name, notebook := range notebooks {
		path := filepath.Join(dir, name)
		if err := SaveNotebook(notebook, path); err != nil {
			return err
		}
	}
	
	return nil
}

// ExpectedOutputs provides expected outputs for test validation
var ExpectedOutputs = map[string][]string{
	"simple_cell_1": {"x = 42"},
	"simple_cell_2": {"y = 84"},
	"data_science_stats": {"mean_x", "mean_y", "std_x", "std_y", "correlation"},
	"ml_training_accuracy": {"Training accuracy:", "Test accuracy:"},
}