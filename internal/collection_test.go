package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeEnvironmentVariables_AddsNewVariables(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{"A": "1"},
	}
	env := Environment{
		Variables: map[string]string{"B": "2", "C": "3"},
	}
	coll.MergeEnvironmentVariables(env)
	assert.Equal(t, "1", coll.Variables["A"])
	assert.Equal(t, "2", coll.Variables["B"])
	assert.Equal(t, "3", coll.Variables["C"])
}

func TestMergeEnvironmentVariables_DoesNotOverwriteExisting(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{"A": "1", "B": "orig"},
	}
	env := Environment{
		Variables: map[string]string{"B": "env", "C": "3"},
	}
	coll.MergeEnvironmentVariables(env)
	assert.Equal(t, "1", coll.Variables["A"])
	assert.Equal(t, "orig", coll.Variables["B"])
	assert.Equal(t, "3", coll.Variables["C"])
}

func TestMergeEnvironmentVariables_HandlesNilCollectionVariables(t *testing.T) {
	coll := &Collection{
		Variables: nil,
	}
	env := Environment{
		Variables: map[string]string{"X": "x"},
	}
	coll.MergeEnvironmentVariables(env)
	assert.Equal(t, "x", coll.Variables["X"])
}

func TestMergeEnvironmentVariables_EmptyEnvironment(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{"A": "1"},
	}
	env := Environment{
		Variables: map[string]string{},
	}
	coll.MergeEnvironmentVariables(env)
	assert.Equal(t, 1, len(coll.Variables))
	assert.Equal(t, "1", coll.Variables["A"])
}

func TestEvaluateVariables_BasicReplacement(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{
			"A": "foo",
			"B": "bar {{A}}",
		},
	}
	vars, err := coll.EvaluateVariables()
	assert.Nil(t, err)
	assert.Equal(t, "foo", vars["A"])
	assert.Equal(t, "bar {{A}}", vars["B"])
}

func TestEvaluateVariables_RandomFunction(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{
			"R": "{{$random.hexadecimal(4)}}",
		},
	}
	vars, err := coll.EvaluateVariables()
	assert.Nil(t, err)
	assert.NotEmpty(t, vars["R"])
	assert.Len(t, vars["R"], 4) // Should be 4 hex digits
}

func TestEvaluateVariables_UnsupportedNamespace(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{
			"X": "{{$foo.bar(1)}}",
		},
	}
	_, err := coll.EvaluateVariables()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsupported namespace")
}

func TestEvaluateVariables_ErrorInFunction(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{
			"X": "{{$random.unknown(1)}}",
		},
	}
	_, err := coll.EvaluateVariables()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in variable")
}

func TestEvaluateVariables_EmptyVariables(t *testing.T) {
	coll := &Collection{
		Variables: map[string]string{},
	}
	vars, err := coll.EvaluateVariables()
	assert.Nil(t, err)
	assert.Empty(t, vars)
}
