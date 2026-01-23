package edgecases

import (
	"math/rand/v2"
	"testing"
)

func TestSelectTagsToOmit(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	tags := SelectTagsToOmit(rng, 3)
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags to omit, got %d", len(tags))
	}
	// Verify all returned tags are in OptionalTags
	for _, tag := range tags {
		found := false
		for _, opt := range OptionalTags {
			if opt == tag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Tag %s not in OptionalTags", tag)
		}
	}
}
