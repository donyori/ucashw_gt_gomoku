package main

import "testing"

func TestPositionParseAndFormatString(t *testing.T) {
	for p := MinPosition; p <= MaxPosition; p++ {
		s := p.String()
		tmp, err := ParsePosition(s)
		if err != nil {
			t.Fatal(err)
		}
		if p != tmp {
			t.Errorf("%v != %v", p, tmp)
		}
	}
	p := InvalidPosition
	s := p.String()
	tmp, err := ParsePosition(s)
	if err != nil {
		t.Fatal(err)
	}
	if p != tmp {
		t.Errorf("%v != %v", p, tmp)
	}
}
