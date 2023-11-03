package v1

import (
	"testing"
)

func TestGetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight(t *testing.T) {
	acqRghtWithLowestPrice := Pair{
		Key: 1,
		Value: Contract{
			Memory: 2,
			Node:   3,
		},
	}
	deltaMemory := 5.0
	deltaNode := 6.0

	result := GetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight(acqRghtWithLowestPrice, deltaMemory, deltaNode)
	expectedResult := 3 // ceil(5/2)

	if result != expectedResult {
		t.Errorf("Expected %d units, but got %d", expectedResult, result)
	}

	// Test with deltaMemory = 0
	deltaMemory = 0
	deltaNode = 6.0
	result = GetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight(acqRghtWithLowestPrice, deltaMemory, deltaNode)
	if result != 2 {
		t.Errorf("Expected 0 units, but got %d", result)
	}
}

func TestGetHighestContractByAmount(t *testing.T) {
	acqRights := map[int]Contract{
		1: {Amount: 100},
		2: {Amount: 200},
		3: {Amount: 50},
	}

	result := GetHighestContractByAmount(acqRights)

	// Sort the expected results based on contract amounts
	expectedResult := []int{2, 1, 3}

	for i, r := range result {
		if r.Key != expectedResult[i] {
			t.Errorf("Expected contract with key %d, but got %d", expectedResult[i], r.Key)
		}
	}
}

func TestGetNumberOfAcquiredRightByOptimizingCost(t *testing.T) {
	acqRights := map[int]Contract{
		1: {Amount: 100, Memory: 1, Node: 2},
		2: {Amount: 200, Memory: 2, Node: 3},
		3: {Amount: 50, Memory: 0.5, Node: 1},
	}
	delta := [1][2]float64{{10, 15}}

	result := GetNumberOfAcquiredRightByOptimizingCost(acqRights, delta)

	// Add assertions to validate the result based on expected outcomes
	// You need to prepare expected results based on the provided data

	if len(result) == 0 {
		t.Error("No contract sets returned")
	}
}

func TestIntegration(t *testing.T) {
	t.Run("TestGetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight", TestGetMaxmimumRequiredContractValueFromLowestPricedAcquiredRight)
	t.Run("TestGetHighestContractByAmount", TestGetHighestContractByAmount)
	t.Run("TestGetNumberOfAcquiredRightByOptimizingCost", TestGetNumberOfAcquiredRightByOptimizingCost)
}
