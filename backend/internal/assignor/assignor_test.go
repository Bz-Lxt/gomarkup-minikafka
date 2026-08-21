package assignor

import (
	"reflect"
	"testing"
)

func TestRoundRobinCoversAll(t *testing.T) {
	m := RoundRobin([]string{"c2", "c1"}, 4)
	if !reflect.DeepEqual(m["c1"], []int{0, 2}) || !reflect.DeepEqual(m["c2"], []int{1, 3}) {
		t.Fatalf("%v", m)
	}
}

func TestRangeConsecutive(t *testing.T) {
	m := Range([]string{"b", "a"}, 5)
	if !reflect.DeepEqual(m["a"], []int{0, 1, 2}) || !reflect.DeepEqual(m["b"], []int{3, 4}) {
		t.Fatalf("%v", m)
	}
}

func TestForMember(t *testing.T) {
	got := ForMember(StrategyRoundRobin, []string{"only"}, 3, "only")
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("%v", got)
	}
}
