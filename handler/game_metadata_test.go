package handler

import (
	"encoding/json"
	"testing"
)

func TestMetadataAndRequirements(t *testing.T) {
	m, err := parseGameMetadata([]byte(`{"730":{"success":true,"data":{"developers":["Valve"],"is_free":true,"pc_requirements":{"minimum":"<strong>Minimum:</strong><br><ul><li>RAM: 8 GB</li></ul>"},"screenshots":[{"path_full":"https://shared.fastly.steamstatic.com/store_item_assets/shot.jpg"}]}}}`), "730")
	if err != nil || m.IsFree == nil || !*m.IsFree || len(m.Developers) != 1 || len(m.Screenshots) != 1 {
		t.Fatalf("metadata: %#v %v", m, err)
	}
	if got := requirementText(m.PC).Minimum; got != "Minimum:\nRAM: 8 GB" {
		t.Fatalf("requirements: %q", got)
	}
	if requirementText(json.RawMessage("[]")).Minimum != "" {
		t.Fatal("array should be empty")
	}
	if steamImageURL("https://evil.example/image.jpg") != "" {
		t.Fatal("untrusted image")
	}
}
