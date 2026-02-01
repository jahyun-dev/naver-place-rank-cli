package main

import "testing"

const sampleHTML = `
<html>
  <body>
    <div class="Ryr1F" id="_pcmap_list_scroll_container">
      <ul>
        <li>
          <span class="OErwL">ad</span>
          <span class="place_bluelink">Ad Place</span>
        </li>
        <li>
          <span class="place_bluelink">First Place</span>
        </li>
        <li>
          <span class="place_bluelink">Second Place</span>
        </li>
      </ul>
    </div>
  </body>
</html>
`

const sampleHTMLWithBadges = `
<html>
  <body>
    <ul>
      <li class="VLTHu OW9LQ">
        <a class="place_bluelink U70Fj k4f_J">
          <span class="YwYLL">플로리에 케이크</span>
          <span class="urQl1">네이버페이</span>
          <span class="urQl1">톡톡</span>
          <span class="YzBgS">케이크전문</span>
        </a>
      </li>
    </ul>
  </body>
</html>
`

func TestFindRankPartialMatch(t *testing.T) {
	rank, matched, items, scanned, err := findRankInHTML([]byte(sampleHTML), "Second", MatchPartial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rank != 2 {
		t.Fatalf("expected rank 2, got %d", rank)
	}
	if matched != "Second Place" {
		t.Fatalf("expected matched name, got %q", matched)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Rank != 1 || items[0].Name != "First Place" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].Rank != 2 || items[1].Name != "Second Place" {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
	if scanned != 2 {
		t.Fatalf("expected scanned 2, got %d", scanned)
	}
}

func TestFindRankExactMatch(t *testing.T) {
	rank, matched, items, _, err := findRankInHTML([]byte(sampleHTML), "Second Place", MatchExact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rank != 2 {
		t.Fatalf("expected rank 2, got %d", rank)
	}
	if matched != "Second Place" {
		t.Fatalf("expected matched name, got %q", matched)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestFindRankNotFound(t *testing.T) {
	rank, matched, items, scanned, err := findRankInHTML([]byte(sampleHTML), "Missing", MatchPartial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rank != -1 {
		t.Fatalf("expected rank -1, got %d", rank)
	}
	if matched != "" {
		t.Fatalf("expected empty matched name, got %q", matched)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if scanned != 2 {
		t.Fatalf("expected scanned 2, got %d", scanned)
	}
}

func TestExtractShopNameWithBadges(t *testing.T) {
	rank, matched, items, scanned, err := findRankInHTML([]byte(sampleHTMLWithBadges), "플로리에", MatchPartial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rank != 1 {
		t.Fatalf("expected rank 1, got %d", rank)
	}
	if matched != "플로리에 케이크" {
		t.Fatalf("expected matched name to be shop name only, got %q", matched)
	}
	if len(items) != 1 || items[0].Name != "플로리에 케이크" || items[0].Rank != 1 {
		t.Fatalf("unexpected items: %+v", items)
	}
	if scanned != 1 {
		t.Fatalf("expected scanned 1, got %d", scanned)
	}
}
