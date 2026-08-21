package assignor_test

import (
	"reflect"
	"testing"

	"minikafka/internal/assignor"
)

func TestAssignPreservesMemberOrder(t *testing.T) {
	members := []string{"worker-c", "worker-a", "worker-b"}
	want := append([]string(nil), members...)

	assignment := assignor.Assign(assignor.StrategyRoundRobin, members, 6)
	if len(assignment) != len(members) {
		t.Fatalf("assignment has %d members, want %d", len(assignment), len(members))
	}
	if !reflect.DeepEqual(members, want) {
		t.Fatalf("Assign changed caller's member order: got %v, want %v", members, want)
	}
}
