package util

import (
	"regexp"
)

// RegexpNamedCaptures Helper function for getting named captures out into a
// map in a sane way from a multiline regex i.e.
// [
// 	{
// 		"capture1": "hello",
// 		"capture2": "dylan"
// 	},
// 	{
// 		"capture1": "hello",
// 		"capture2": "francis"
// 	},
// ]
func RegexpNamedCaptures(r *regexp.Regexp, s string) []map[string]string {
	matches := r.FindAllStringSubmatch(s, 65535)
	var results []map[string]string
	// Loop over all of the times the regexp matched
	for _, match := range matches {
		result := make(map[string]string)
		// Loop over all the captures and add them to the result map
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		// Append the result
		results = append(results, result)
	}
	return results
}
