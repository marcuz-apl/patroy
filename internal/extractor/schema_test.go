package extractor

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractCustomSchema(t *testing.T) {
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head><title>Product Catalog</title></head>
	<body>
		<h1 class="product-title">Ultra Monitor 4K</h1>
		<div class="price">$499.99</div>
		<img class="hero-image" src="https://example.com/img.jpg" alt="Monitor" />
		<ul class="specs">
			<li class="spec-item">3840x2160</li>
			<li class="spec-item">144Hz</li>
			<li class="spec-item">IPS Panel</li>
		</ul>
	</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	schema := map[string]interface{}{
		"title":     "h1.product-title",
		"price":     ".price",
		"specs":     []interface{}{".spec-item"},
		"image_url": map[string]interface{}{"selector": "img.hero-image", "attr": "src"},
	}

	data, err := ExtractCustomSchema(doc, schema)
	if err != nil {
		t.Fatalf("ExtractCustomSchema failed: %v", err)
	}

	if data["title"] != "Ultra Monitor 4K" {
		t.Errorf("expected title 'Ultra Monitor 4K', got %v", data["title"])
	}
	if data["price"] != "$499.99" {
		t.Errorf("expected price '$499.99', got %v", data["price"])
	}
	if data["image_url"] != "https://example.com/img.jpg" {
		t.Errorf("expected image_url 'https://example.com/img.jpg', got %v", data["image_url"])
	}

	specs, ok := data["specs"].([]string)
	if !ok || len(specs) != 3 {
		t.Fatalf("expected 3 specs items, got %v", data["specs"])
	}
	if specs[0] != "3840x2160" || specs[1] != "144Hz" || specs[2] != "IPS Panel" {
		t.Errorf("unexpected specs content: %v", specs)
	}
}
