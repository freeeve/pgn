package pgn

import "testing"

func TestParseUCI(t *testing.T) {
	tests := []struct {
		uci     string
		wantErr bool
		from    Square
		to      Square
		promo   PromoPiece
	}{
		{"e2e4", false, 12, 28, NoPromo},
		{"g1f3", false, 6, 21, NoPromo},
		{"e7e8q", false, 52, 60, PromoQueen},
		{"a7a8r", false, 48, 56, PromoRook},
		{"b7b8b", false, 49, 57, PromoBishop},
		{"c7c8n", false, 50, 58, PromoKnight},
		{"e7e8Q", false, 52, 60, PromoQueen}, // uppercase promo
		{"a1h8", false, 0, 63, NoPromo},
		{"h1a8", false, 7, 56, NoPromo},
		// Error cases
		{"e2", true, 0, 0, NoPromo},       // too short
		{"e2e4e5", true, 0, 0, NoPromo},   // too long
		{"x2e4", true, 0, 0, NoPromo},     // invalid from file
		{"e9e4", true, 0, 0, NoPromo},     // invalid from rank
		{"e2x4", true, 0, 0, NoPromo},     // invalid to file
		{"e2e9", true, 0, 0, NoPromo},     // invalid to rank
		{"e7e8x", true, 0, 0, NoPromo},    // invalid promo piece
	}

	for _, tt := range tests {
		t.Run(tt.uci, func(t *testing.T) {
			mv, err := ParseUCI(tt.uci)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseUCI(%q) expected error, got nil", tt.uci)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseUCI(%q) error: %v", tt.uci, err)
			}
			if mv.From != tt.from {
				t.Errorf("ParseUCI(%q).From = %d, want %d", tt.uci, mv.From, tt.from)
			}
			if mv.To != tt.to {
				t.Errorf("ParseUCI(%q).To = %d, want %d", tt.uci, mv.To, tt.to)
			}
			if mv.Promo != tt.promo {
				t.Errorf("ParseUCI(%q).Promo = %d, want %d", tt.uci, mv.Promo, tt.promo)
			}
		})
	}
}

func TestParseUCI_RoundTrip(t *testing.T) {
	// Parse and then convert back to string should give same result
	moves := []string{"e2e4", "g1f3", "e7e8q", "a7a8r", "a1h8"}
	for _, uci := range moves {
		mv, err := ParseUCI(uci)
		if err != nil {
			t.Fatalf("ParseUCI(%q) error: %v", uci, err)
		}
		got := mv.String()
		if got != uci {
			t.Errorf("Round-trip: ParseUCI(%q).String() = %q", uci, got)
		}
	}
}
