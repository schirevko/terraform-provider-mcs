package valid

import "testing"

func TestAvailabilityZone(t *testing.T) {
	tests := map[string]struct {
		zone string
		err  error
	}{
		// ok
		"dp1": {zone: "dp1", err: nil},
		"DP1": {zone: "DP1", err: nil},
		"ms1": {zone: "ms1", err: nil},
		// invalid zone
		"ms2":         {zone: "ms2", err: ErrInvalidAvailabilityZone},
		"empty value": {zone: "", err: ErrInvalidAvailabilityZone},
	}

	for name := range tests {
		tt := tests[name]
		t.Run(name, func(t *testing.T) {
			if err := AvailabilityZone(tt.zone); err != tt.err {
				t.Errorf("err got=%s; want=%s", err, tt.err)
			}
		})
	}
}
