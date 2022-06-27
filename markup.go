package markupstring

import (
	"errors"
	"regexp"
	"strings"
)

type Markup struct {
	Name      string
	Value     interface{}
	Reference string
}

type MarkupEntry struct {
	Text       string
	Properties []Markup
	Parsed     string
}

type MarkupParser struct {
	HandleErrors bool
	Entries      []MarkupEntry
	LeftString   string // at least the part of the string, until the first markup
}

type MarkupRunner struct {
	Exec func(mk Markup, current string) string
}

type Runner struct {
	HandleErrors      bool
	runners           map[string]MarkupRunner
	leftBracket       byte
	rightRBracket     byte
	level1Sep         byte
	level2Sep         byte
	UseCache          bool
	markupEntrieCache map[string][]MarkupEntry
	markupCache       map[string]Markup
}

func NewRunner(handleErrors bool, bracketsAndSepByChar string) (*Runner, error) {

	if len(bracketsAndSepByChar) != 4 {
		return &Runner{}, errors.New("no valid bracket and seperatot definition found. you need to set these up like <>:;")
	}

	return &Runner{
		runners:           make(map[string]MarkupRunner),
		markupEntrieCache: make(map[string][]MarkupEntry),
		leftBracket:       bracketsAndSepByChar[0],
		rightRBracket:     bracketsAndSepByChar[1],
		level1Sep:         bracketsAndSepByChar[2],
		level2Sep:         bracketsAndSepByChar[3],
		HandleErrors:      handleErrors,
	}, nil
}

// ParseAll is actually the main handler to run all
// parsings
func (mp *MarkupParser) ParseAll(runner Runner) (string, error) {
	current := ""
	var results []string
	mp.HandleErrors = runner.HandleErrors
	for _, me := range mp.Entries {
		if cur, err := me.ExecParse(runner, current, mp.HandleErrors); err != nil {
			if mp.HandleErrors {
				return cur, err
			}
		} else {
			results = append(results, cur)
			current = cur
		}

	}
	return mp.LeftString + strings.Join(results, ""), nil
}

func (mark *MarkupEntry) ExecParse(runner Runner, current string, stopOnErrors bool) (string, error) {
	var results []string
	for _, mup := range mark.Properties {

		if runner, exists := runner.runners[mup.Name]; exists {
			add := runner.Exec(mup, current)
			results = append(results, add)
		} else {
			if stopOnErrors {
				return "", errors.New("undefined propertie: " + mup.Name)
			} else {
				results = append(results, "[ERROR] unknown markup")
			}
		}
	}
	return strings.Join(results, ""), nil
}

func (r *Runner) AddRunner(responsible string, runner MarkupRunner) error {
	if _, exists := r.runners[responsible]; exists {
		return errors.New("runner already exists for " + responsible)
	}
	r.runners[responsible] = runner
	return nil
}

func (r *Runner) getMarkups(orig string, ref string) []Markup {
	var marks []Markup
	orig = strings.TrimLeft(orig, string(r.leftBracket))
	orig = strings.TrimRight(orig, string(r.rightRBracket))
	parts := strings.Split(orig, string(r.level1Sep))
	for _, pstr := range parts {
		pParts := strings.Split(pstr, string(r.level2Sep))
		var m Markup
		m.Name = pParts[0]
		m.Reference = ref
		if len(pParts) > 1 {
			m.Value = pParts[1]
		}
		marks = append(marks, m)
	}
	return marks
}

// ParseMarkup is parsing the origin string.
// what means it extracts all markups in the
// the string and then create an assignemnt
// what markup is responsible for which part
// ot the text
func (r *Runner) ParseMarkup(orig string) MarkupParser {
	var parsed MarkupParser

	// we have maybe already did this work already.
	// if so, then we can use the same result, if we cachaed them.
	if r.UseCache {
		if entr, exists := r.markupEntrieCache[orig]; exists {
			parsed.Entries = entr
			return parsed
		}
	}
	// first extract the markups, and iterate over them, if we found some
	if markups, found := r.splitByMarks(orig); found {
		allMk := len(markups)
		// simple iterate over all markups
		for mkIndex, mk := range markups {
			// get the postion of the markup in the orogin string
			if indxStart := strings.Index(orig, mk); indxStart >= 0 {
				//at the first markup, we need also keep the left side of the origin string
				if mkIndex == 0 {
					parsed.LeftString = orig[:indxStart]
				}
				// the length of the markup string.
				// we need them for some calculations depending the string indecies
				offset := len(mk)
				text := ""
				// the right part of the string
				// would have no markup, so all the
				// rest is the content for the last markup
				// any other part of the text is handled
				// in the else part
				if mkIndex+1 < allMk {
					if end := strings.Index(orig[indxStart+offset:], string(r.leftBracket)); end >= 0 {
						text = orig[indxStart+offset : indxStart+end+offset]
					}
				} else {
					text = orig[indxStart+offset:]
				}
				// create a new markup entry.
				// do not overlook r.getMarkups because
				// there we parsing the markup sub enties
				var entry MarkupEntry = MarkupEntry{
					Text:       text,
					Properties: r.getMarkups(mk, text),
				}
				parsed.Entries = append(parsed.Entries, entry)
			}
		}
	}
	// for cach usage, we keep the result
	// assigned to the origin string.
	if r.UseCache {
		r.markupEntrieCache[orig] = parsed.Entries
	}
	return parsed
}

// splitByMarks extract all markups parts
func (r *Runner) splitByMarks(orig string) ([]string, bool) {

	var result []string
	found := false
	cmpStr := string(r.leftBracket) + "[^" + string(r.rightRBracket) + "]+" + string(r.rightRBracket)
	re := regexp.MustCompile(cmpStr)
	newStrs := re.FindAllString(orig, -1)
	for _, s := range newStrs {
		found = true
		result = append(result, s)

	}
	return result, found
}
