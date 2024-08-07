package plugin

import (
	"fmt"
	"regexp"
	"sort"
)

// tagRegex is the regex to extract the tag from a test name.
// The 'tag' is the first bracket content from a test name.
const tagRegex = `^\[([a-zA-Z0-9-]*)\]`

// TestTags stores the test tags map with it's counter.
// The test tag is the work extracted from the first bracket from a test name.
// Example test name: '[sig-provider] test name' the 'sig-provider' is the tag.
type TestTags map[string]int

// NewTestTagsEmpty creates the TestTags with a specific size, to be populated later.
func NewTestTagsEmpty(size int) TestTags {
	tt := make(TestTags, size)
	tt["total"] = 0
	return tt
}

// NewTestTags creates the TestTags populating the tag values and counters.
func NewTestTags(tests []*string) TestTags {
	tt := make(TestTags, len(tests))
	tt["total"] = 0
	tt.addBatch(tests)
	return tt
}

// Add extracts tags from test name, store, and increment the counter.
func (tt TestTags) Add(test *string) {
	reT := regexp.MustCompile(tagRegex)
	match := reT.FindStringSubmatch(*test)
	if len(match) > 0 {
		if _, ok := tt[match[1]]; !ok {
			tt[match[1]] = 1
		} else {
			tt[match[1]] += 1
		}
	}
	tt["total"] += 1
}

// AddBatch receive a list of test name (string slice) and stores it.
func (tt TestTags) addBatch(kn []*string) {
	for _, test := range kn {
		tt.Add(test)
	}
}

// SortRev creates a rank of tags.
func (tt TestTags) sortRev() []SortedDict {
	tags := make(SortedList, len(tt))
	i := 0
	for k, v := range tt {
		tags[i] = SortedDict{k, v}
		i++
	}
	sort.Sort(sort.Reverse(tags))
	return tags
}

// ShowSorted return an string with the rank of tags.
func (tt TestTags) ShowSorted() string {
	tags := tt.sortRev()
	msg := ""
	for _, k := range tags {
		if k.Key == "total" {
			msg = fmt.Sprintf("[%v=%v]", k.Key, k.Value)
			continue
		}
		msg = fmt.Sprintf("%s [%v=%s]", msg, k.Key, UtilsCalcPercStr(int64(k.Value), int64(tt["total"])))
	}
	return msg
}

// calcPercStr receives the numerator and denominator and return the numerator and percentage as string.
func UtilsCalcPercStr(num, den int64) string {
	return fmt.Sprintf("%d (%.2f%%)", num, (float64(num)/float64(den))*100)
}
