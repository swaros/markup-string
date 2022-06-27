package markupstring_test

import (
	"fmt"
	"strings"
	"testing"

	markupstring "github.com/swaros/markup-string"
)

func TestInvalidCreate(t *testing.T) {
	_, err := markupstring.NewRunner(true, "hello world")

	if err == nil {
		t.Error("Error should happen, if we provide wrong init string")
	}
}

func TestRegularCreate(t *testing.T) {
	runParser, err := markupstring.NewRunner(true, "<>;=")

	if err != nil {
		t.Error(err)
	}
	parseIsExeucted := 0
	runParser.AddRunner("parse", markupstring.MarkupRunner{
		Exec: func(mk markupstring.Markup, current string) string {
			parseIsExeucted++
			return "[P]" + strings.ToUpper(mk.Reference) + "[/P]"
		},
	})

	runParser.AddRunner("mod", markupstring.MarkupRunner{
		Exec: func(mk markupstring.Markup, current string) string {
			return "[M]" + strings.ToLower(mk.Reference) + "[/M]"
		},
	})

	toParse := "1.Regular Start...<parse>2.other string<mod>...3.the rest"

	parser := runParser.ParseMarkup(toParse)
	if result, rErr := parser.ParseAll(*runParser); rErr != nil {
		t.Error(rErr)
	} else {
		fmt.Println(result)
		if result != "1.Regular Start...[P]2.OTHER STRING[/P][M]...3.the rest[/M]" {
			t.Error("unexecpected result:[", result, "]")
		}

		if parseIsExeucted != 1 {
			t.Error("parse is not executed once. ", parseIsExeucted)
		}
	}

}

func TestRegularCreateDoubleMarkups(t *testing.T) {
	runParser, err := markupstring.NewRunner(true, "<>;=")

	if err != nil {
		t.Error(err)
	}
	parseIsExeucted := 0
	runParser.AddRunner("parse", markupstring.MarkupRunner{
		Exec: func(mk markupstring.Markup, current string) string {
			parseIsExeucted++
			return "[P]" + strings.ToUpper(mk.Reference) + "[/P]"
		},
	})

	runParser.AddRunner("mod", markupstring.MarkupRunner{
		Exec: func(mk markupstring.Markup, current string) string {
			return "[M]" + strings.ToLower(mk.Reference) + "[/M]"
		},
	})

	toParse := "1.Regular Start...<parse><mod>2.other string<mod>...3.the rest"

	parser := runParser.ParseMarkup(toParse)
	if result, rErr := parser.ParseAll(*runParser); rErr != nil {
		t.Error(rErr)
	} else {
		fmt.Println(result)
		if result != "1.Regular Start...[P][/P][M]2.other string[/M][M]...3.the rest[/M]" {
			t.Error("unexecpected result:[", result, "]")
		}

		if parseIsExeucted != 1 {
			t.Error("parse is not executed once. ", parseIsExeucted)
		}
	}

}
