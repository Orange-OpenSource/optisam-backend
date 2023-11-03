package v1

import (
	"testing"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func TestGetChildEquipmentsByParentType(t *testing.T) {
	// Example equipment types for testing
	eqTypes := []*repo.EquipmentType{
		{Type: "Child1", ParentType: "ParentA"},
		{Type: "Child2", ParentType: "ParentA"},
		{Type: "Child3", ParentType: "ParentB"},
		{Type: "Child4", ParentType: "ParentA"},
		{Type: "Child5", ParentType: "ParentC"},
	}

	tests := []struct {
		parentType  string
		expected    []string
		description string
	}{
		{"ParentA", []string{"Child1", "Child2", "Child4"}, "Matching parent type ParentA"},
		{"ParentB", []string{"Child3"}, "Matching parent type ParentB"},
		{"ParentC", []string{"Child5"}, "Matching parent type ParentC"},
		{"ParentD", []string{}, "No matching parent type"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := getChildEquipmentsByParentType(test.parentType, eqTypes)
			if len(result) != len(test.expected) {
				t.Errorf("For parent type %s, expected %v child equipment types but got %v", test.parentType, len(test.expected), len(result))
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("For parent type %s, expected child equipment type %s but got %s", test.parentType, test.expected[i], result[i])
				}
			}
		})
	}
}
