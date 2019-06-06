package main

import (
	"testing"
)

func TestSanitarizeString(t *testing.T) {
	healthyString := "ThisIsAHealthyLittleString"
	sanitarized := sanitarizeString(healthyString)

	if healthyString != sanitarized {
		t.Errorf("Nothing should have been removed, " +
			"healty is \"%s\", sanitarized was \"%s\"\n", healthyString, sanitarized)
	}

	stringWithSpace := "This Is A String With Space"
	expected := "ThisIsAStringWithSpace"
	sanitarized = sanitarizeString(stringWithSpace)

	if sanitarized != expected {
		t.Errorf("didn't remove spaces, expected \"%s\", got \"%s\"", expected, sanitarized)
	}

	stringWithDots := "This.Is.A.String.With.Dots"
	expected = "ThisIsAStringWithDots"
	sanitarized = sanitarizeString(stringWithDots)

	if sanitarized != expected {
		t.Errorf("didn't remove dots, expected \"%s\", got \"%s\"", expected, sanitarized)
	}

	stringWithSlashes := "This/Is/A/String/With/Slash"
	expected  = "ThisIsAStringWithSlash"
	sanitarized = sanitarizeString(stringWithSlashes)

	if sanitarized != expected {
		t.Errorf("didn't remove slashes, expected \"%s\", got \"%s\"", expected, sanitarized)
	}

	stringWithAll := "This Is A ..String/With./Lots Of Bad ...././../Stuff"
	expected = "ThisIsAStringWithLotsOfBadStuff"
	sanitarized = sanitarizeString(stringWithAll)

	if sanitarized != expected {
		t.Errorf("didnt sanitarize at all, expected \"%s\", got \"%s\"", expected, sanitarized)
	}
}
