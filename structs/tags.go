package structs

import "strings"

// tagOptions contains a slice of tag options
type tagOptions []string

// HasOmitnested returns true if omitnested is available in tagOptions
func (t tagOptions) HasOmitnested() bool { return t.Has("omitnested") }

// HasOmitempty returns true if omitempty is available in tagOptions
func (t tagOptions) HasOmitempty() bool { return t.Has("omitempty") }

// HasString returns true if string is available in tagOptions
func (t tagOptions) HasString() bool { return t.Has("string") }

// HasFlatten returns true if flatten is available in tagOptions
func (t tagOptions) HasFlatten() bool { return t.Has("flatten") }

// Has returns true if the given option is available in tagOptions
func (t tagOptions) Has(opt string) bool {
	for _, tagOpt := range t {
		if tagOpt == opt {
			return true
		}
	}

	return false
}

// parseTag splits a struct field's tag into its name and a list of options
// which comes after a name. A tag is in the form of: "name,option1,option2".
// The name can be neglected.
func parseTag(tag string) (string, tagOptions) {
	// tag is one of followings:
	// ""
	// "name"
	// "name,opt"
	// "name,opt,opt2"
	// ",opt"
	res := strings.Split(tag, ",")
	return res[0], res[1:]
}
