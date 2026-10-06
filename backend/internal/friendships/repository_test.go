package friendships

import "testing"

func TestPairCanonicalizesOrder(t *testing.T) {
	const first = "00000000-0000-4000-8000-000000000001"
	const second = "ffffffff-ffff-4fff-8fff-ffffffffffff"

	a, b := pair("FFFFFFFF-FFFF-4FFF-8FFF-FFFFFFFFFFFF", first)
	if a != first || b != second {
		t.Fatalf("pair() = (%q, %q), want (%q, %q)", a, b, first, second)
	}
}
