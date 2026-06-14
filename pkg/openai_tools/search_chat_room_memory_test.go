package openaitools

import "testing"

func TestHumanMemoryLabels(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "confidence", got: humanConfidence(86), want: "可信度86"},
		{name: "alias type", got: humanAliasType("observed_call_name"), want: "群里常用叫法"},
		{name: "fact predicate", got: humanFactPredicate("occupation"), want: "职业"},
		{name: "relation type", got: humanRelationType("joke_partner"), want: "经常开玩笑"},
		{name: "direction", got: humanRelationDirection("directed"), want: "单向关系"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, tt.got)
			}
		})
	}
}
