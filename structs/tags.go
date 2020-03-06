package structs

import "strings"

// tagOptions contains a slice of tag options
type tagOptions struct {
	TagOptions []string
	Option     *Option
}

// OmitNested returns true if omitnested is available in tagOptions
func (t tagOptions) OmitNested() bool { return t.Option.OmitNested || t.Has("omitnested") }

// OmitEmpty returns true if omitempty is available in tagOptions
func (t tagOptions) OmitEmpty() bool { return t.Option.OmitEmpty || t.Has("omitempty") }

// Stringer returns true if string is available in tagOptions
func (t tagOptions) Stringer() bool { return t.Option.Stringer || t.Has("string") }

// Flatten returns true if flatten is available in tagOptions
func (t tagOptions) Flatten() bool { return t.Option.Flatten || t.Has("flatten") }

// Has returns true if the given option is available in tagOptions
func (t tagOptions) Has(opt string) bool {
	for _, tagOpt := range t.TagOptions {
		if tagOpt == opt {
			return true
		}
	}

	return false
}

// parseTag splits a struct field's tag into its name and a list of options
// which comes after a name. A tag is in the form of: "name,option1,option2".
// The name can be neglected.
func parseTag(option *Option, tag string) (string, tagOptions) {
	// tag is one of followings:
	// ""
	// "name"
	// "name,opt"
	// "name,opt,opt2"
	// ",opt"
	res := strings.Split(tag, ",")
	return res[0], tagOptions{TagOptions: res[1:], Option: option}
}
