package markupstring_test

import (
	"fmt"
	"strconv"
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
	if result, rErr := parser.ParseAll(runParser); rErr != nil {
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
	if result, rErr := parser.ParseAll(runParser); rErr != nil {
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

	runParser.DisableParsing()
	cleanTxt, _ := runParser.Parse(toParse)

	if cleanTxt != "1.Regular Start...2.other string...3.the rest" {
		t.Error("unexpected string:", cleanTxt)
	}
}

func TestMarkupArgs(t *testing.T) {
	runParser, err := markupstring.NewRunner(true, "{};=")
	if err != nil {
		t.Error(err)
	}

	runParser.AddRunner("left", markupstring.MarkupRunner{
		Exec: func(mk markupstring.Markup, current string) string {
			var len int
			len, _ = strconv.Atoi(mk.Value.(string))
			return current[len:]
		},
	})

	runParser.AddRunner("right", markupstring.MarkupRunner{
		Exec: func(mk markupstring.Markup, current string) string {
			var tlen int
			tlen, _ = strconv.Atoi(mk.Value.(string))

			return current[:len(current)-tlen]
		},
	})

	source := "check one. {left=4;right=5}abcdefghijklmnopqrstuvwxyz"
	if res, err := runParser.Parse(source); err != nil {
		t.Error(err)
	} else {
		if !strings.EqualFold(res, "check one. efghijklmnopqrstu") {
			t.Error("unexpected string:", res)
		}
	}
}

func TestStopOnError(t *testing.T) {
	if runner, err := markupstring.NewRunner(true, "<>;="); err != nil {
		t.Error(err)
	} else {
		if _, err2 := runner.Parse("<check>it out"); err2 == nil {
			t.Error("error should be triggered")
		} else {

			if err2.Error() != "undefined propertie: check" {
				t.Error("unexpected error message:", err2.Error())
			}
		}
	}
}
