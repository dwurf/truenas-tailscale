package truenas

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestApp_UnmarshalJSON(t *testing.T) {
	t.Run("can unmarshal", func(t *testing.T) {
		var app App
		if err := json.Unmarshal([]byte(`{"name":"minio", "state": "RUNNING", "portals": {"webUI": "http://127.0.0.1:9002"}}`), &app); err != nil {
			t.Fatalf("unmarshal failed: %s", err)
		}

		if app.Name != "minio" {
			t.Errorf("Unexpected name, want: %s, got: %s", "minio", app.Name)
		}
		if app.State != StateRunning {
			t.Errorf("Unexpected state, want: %s, got: %s", StateRunning, app.State)
		}
		if len(app.Portals) != 1 {
			t.Fatal("Got zero portals, expected 1")
		}
		portal, err := url.Parse("http://127.0.0.1:9002")
		if err != nil {
			t.Fatal("Invalid URL")
		}
		if *app.Portals[0] != *portal {
			t.Errorf("Unexpected portal, want %v, got %v", *portal, *app.Portals[0])
		}
	})
}
