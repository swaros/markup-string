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
	disabled          bool
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
func (mp *MarkupParser) ParseAll(runner *Runner) (string, error) {
	current := ""
	var results []string
	mp.HandleErrors = runner.HandleErrors
	for _, me := range mp.Entries {
		if runner.disabled {
			results = append(results, me.Text)
		} else {

			if cur, err := me.ExecParse(runner, current, mp.HandleErrors); err != nil {
				if mp.HandleErrors {
					return cur, err
				}
			} else {
				results = append(results, cur)
				current = cur
			}
		}
	}
	return mp.LeftString + strings.Join(results, ""), nil
}

func (r *Runner) Parse(str string) (string, error) {
	markParser := r.ParseMarkup(str)
	return markParser.ParseAll(r)
}

func (mark *MarkupEntry) ExecParse(runner *Runner, current string, stopOnErrors bool) (string, error) {
	output := ""
	for _, mup := range mark.Properties {

		if runner, exists := runner.runners[mup.Name]; exists {
			output = runner.Exec(mup, output)
		} else {
			if stopOnErrors {
				return "", errors.New("undefined propertie: " + mup.Name)
			}
		}
	}
	return output, nil
}

// AddRunner applies the logic to an markup
func (r *Runner) AddRunner(responsible string, runner MarkupRunner) error {
	if _, exists := r.runners[responsible]; exists { // we do not allow overwrite existing runners
		return errors.New("runner already exists for " + responsible)
	}
	r.runners[responsible] = runner
	return nil
}

func (r *Runner) DisableParsing() *Runner {
	r.disabled = true
	return r
}

func (r *Runner) EnableParsing() *Runner {
	r.disabled = false
	return r
}

func (r *Runner) IsDisabled() bool {
	return r.disabled
}

// getMarkups parses the markup string itself to get all sub definitions within the markup
// any markup can contains multiple 'real' markusp,that seperated by the level1 seperator
func (r *Runner) getMarkups(markupAsString string, ref string) []Markup {
	var marks []Markup                                                          // this will be oure result
	markupAsString = strings.TrimLeft(markupAsString, string(r.leftBracket))    // remove the markup strings. on left
	markupAsString = strings.TrimRight(markupAsString, string(r.rightRBracket)) // and right
	parts := strings.Split(markupAsString, string(r.level1Sep))                 // first spits by the level1 seperator. it splits the markups itself
	for _, pstr := range parts {
		pParts := strings.Split(pstr, string(r.level2Sep)) // split again by using the second sperarator. that is usual an assigment like =
		var m Markup
		m.Name = pParts[0]   // the name of the markup is always the left side.
		m.Reference = ref    // this is the context. meaning the string that needs to be woking on
		if len(pParts) > 1 { // if we have different assignements (by split using level2sep), then we assign them too
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

	searchString := orig                               // we need a copy of the origin string, so we can cut them after any search hit
	if markups, found := r.splitByMarks(orig); found { // first extract the markups, and iterate over them, if we found some
		allMk := len(markups)

		for mkIndex, mk := range markups { // simple iterate over all markups

			if indxStart := strings.Index(searchString, mk); indxStart >= 0 { // get the postion of the markup in the orogin string
				if mkIndex == 0 { //at the first markup, we need also keep the left side of the origin string
					parsed.LeftString = searchString[:indxStart] // keep the left side
				}
				offset := len(mk) // the length of the markup string. we need them for some calculations depending the string indecies

				text := ""
				// the right part of the string
				// would have no markup, so all the
				// rest is the content for the last markup
				// any other part of the text is handled
				// in the else part
				if mkIndex+1 < allMk {
					if end := strings.Index(searchString[indxStart+offset:], string(r.leftBracket)); end >= 0 {
						text = searchString[indxStart+offset : indxStart+end+offset]
					}
				} else {
					text = searchString[indxStart+offset:]
				}
				searchString = searchString[indxStart+offset:]
				var entry MarkupEntry = MarkupEntry{ // create a new markup entry.
					Text:       text,
					Properties: r.getMarkups(mk, text), // do not overlook r.getMarkups because there we parsing the markup sub enties
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

// splitByMarks extract all parts of the text, that
// matched to the pattern. that means it starts
// with the char leftBracket, have some content, and ends with
// the rightBracket.
// these two bracket definitions the only nes we used here
func (r *Runner) splitByMarks(orig string) ([]string, bool) {
	var result []string
	found := false
	cmpStr := string(r.leftBracket) + "[^" + string(r.rightRBracket) + "]+" + string(r.rightRBracket) // compose the regex
	re := regexp.MustCompile(cmpStr)                                                                  // compile and panic on bad things happens
	newStrs := re.FindAllString(orig, -1)                                                             // use regex to find all patterns
	for _, s := range newStrs {                                                                       // get all markups
		found = true
		result = append(result, s)

	}
	return result, found
}
