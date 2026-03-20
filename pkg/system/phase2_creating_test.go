package system

import "testing"

func TestValidateAzureSTSParams(t *testing.T) {
	tests := []struct {
		name            string
		clientID        string
		resourcegroupID string
		tenantID        string
		subscriptionID  string
		wantOK          bool
	}{
		{
			name:   "none set",
			wantOK: true,
		},
		{
			name:            "all four non-empty",
			clientID:        "c",
			resourcegroupID: "rg",
			tenantID:        "t",
			subscriptionID:  "s",
			wantOK:          true,
		},
		{
			name:     "only clientID",
			clientID: "c",
			wantOK:   false,
		},
		{
			name:            "three of four",
			clientID:        "c",
			resourcegroupID: "rg",
			tenantID:        "t",
			wantOK:          false,
		},
		{
			name:            "two of four",
			clientID:        "c",
			resourcegroupID: "rg",
			wantOK:          false,
		},
		{
			name:            "whitespace-only strings count as set — all four",
			clientID:        " ",
			resourcegroupID: " ",
			tenantID:        " ",
			subscriptionID:  " ",
			wantOK:          true,
		},
		{
			name:     "one non-empty and rest empty",
			clientID: "x",
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateAzureSTSParams(tt.clientID, tt.resourcegroupID, tt.tenantID, tt.subscriptionID)
			if got != tt.wantOK {
				t.Errorf("validateAzureSTSParams() = %v, want %v", got, tt.wantOK)
			}
		})
	}
}
