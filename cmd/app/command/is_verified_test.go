package command

import (
	"reflect"
	"testing"
)

func TestVerifiedStatus(t *testing.T) {
	tests := []struct {
		name          string
		verifiedLabel map[string]interface{}
		want          map[string]interface{}
	}{
		{
			name:          "label not configured for the project",
			verifiedLabel: nil,
			want: map[string]interface{}{
				"status": "not-configured",
			},
		},
		{
			name: "approved by a +1 vote",
			verifiedLabel: map[string]interface{}{
				"approved": map[string]interface{}{"name": "svc-moab2gerrit"},
				"all": []interface{}{
					map[string]interface{}{"name": "svc-moab2gerrit", "value": float64(1)},
				},
			},
			want: map[string]interface{}{
				"status":   "verified",
				"verified": true,
				"by":       "svc-moab2gerrit",
			},
		},
		{
			name: "rejected by a -1 vote",
			verifiedLabel: map[string]interface{}{
				"rejected": map[string]interface{}{"name": "svc-moab2gerrit"},
				"all": []interface{}{
					map[string]interface{}{"name": "svc-moab2gerrit", "value": float64(-1)},
				},
			},
			want: map[string]interface{}{
				"status":   "rejected",
				"verified": false,
				"by":       "svc-moab2gerrit",
			},
		},
		{
			name: "no votes cast yet",
			verifiedLabel: map[string]interface{}{
				"all": []interface{}{},
			},
			want: map[string]interface{}{
				"status":   "no-score",
				"verified": false,
			},
		},
		{
			name: "only a zero vote recorded",
			verifiedLabel: map[string]interface{}{
				"all": []interface{}{
					map[string]interface{}{"name": "someone", "value": float64(0)},
				},
			},
			want: map[string]interface{}{
				"status":   "no-score",
				"verified": false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := verifiedStatus(test.verifiedLabel)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("verifiedStatus() = %v, want %v", got, test.want)
			}
		})
	}
}
