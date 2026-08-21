package validate

import "testing"

func TestResourceName(t *testing.T) {
	if err := ResourceName("topic", "orders"); err != nil {
		t.Fatal(err)
	}
	if err := ResourceName("topic", ""); err == nil {
		t.Fatal("empty should fail")
	}
	if err := ResourceName("topic", "bad name"); err == nil {
		t.Fatal("space should fail")
	}
}

func TestPartitionCount(t *testing.T) {
	n, err := PartitionCount(0)
	if err != nil || n != 1 {
		t.Fatalf("default %d %v", n, err)
	}
	if _, err := PartitionCount(32); err == nil {
		t.Fatal("32 should fail")
	}
}

func TestLimits(t *testing.T) {
	if ConsumeLimit(0) != 50 || BrowseLimit(999) != 20 {
		t.Fatal("limit defaults")
	}
	if err := ResetTarget("latest"); err != nil {
		t.Fatal(err)
	}
	if err := ResetTarget("middle"); err == nil {
		t.Fatal("middle")
	}
	if err := BatchSize(0); err == nil {
		t.Fatal("empty batch")
	}
}
