package handlers

import "testing"

func TestParseEventTime(t *testing.T) {
	valid := map[string][2]int{
		"10:30": {10, 30},
		"9:05":  {9, 5},
		"00:00": {0, 0},
		"23:59": {23, 59},
	}
	for s, want := range valid {
		got, err := parseEventTime(s)
		if err != nil {
			t.Errorf("parseEventTime(%q) error: %v", s, err)
			continue
		}
		if got.Hour() != want[0] || got.Minute() != want[1] {
			t.Errorf("parseEventTime(%q) = %s, want %02d:%02d", s, got.Format("15:04"), want[0], want[1])
		}
	}

	for _, s := range []string{"", "10", "9:5", "24:00", "10:60", "ab:cd", "10:30:00", " 10:30"} {
		if _, err := parseEventTime(s); err == nil {
			t.Errorf("parseEventTime(%q) must fail", s)
		}
	}
}
