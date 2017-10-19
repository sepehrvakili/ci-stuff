package texter

import (
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

// TagMerger merges actionkit tags with values
type TagMerger interface {
	MergeRep(body string, rep RepInfo) string
	MergeUnknown(body string) string
}

// NewTagMerger is a factory method that creates a new tag merger
func NewTagMerger() TagMerger {
	return CongressTagMerger{}
}

// CongressTagMerger is a basic merger that handles information for
// house of representative members.
type CongressTagMerger struct {
}

// Local file matcher, no need to make this an instance variable.
var matcher = regexp.MustCompile(`[{]{2}(?P<tag>([\w|\s|\.\-_])*)[}]{2}`)

// UnknownTagReplacement constant represents the text that will be replaced
// for any tag that we do not understand.
const UnknownTagReplacement = "unknown-tag"

// UnknownDataTypeReplacement constant represents the text that will be replaced for any
// tag whose data type we do not understand.
const UnknownDataTypeReplacement = "unknown-data-type"

// InvalidFieldNameReplacement represents a programmer error. Please fix the map of
// merge tags to field names.
const InvalidFieldNameReplacement = "invalid-field-name"

// Map of tags to struct field names.
var tags = map[string]string{
	"{{targets.title}}":       "LongTitle",
	"{{targets.full_name}}":   "OfficialName",
	"{{targets.phone}}":       "PhoneNumber",
	"{{targets.last_name}}":   "LastName",
	"{{targets.short_title}}": "Title",
}

// MergeRep merges information from the target entity into the body. The
// entity will be assumed to be a rep info object, but can be anything.
// So long as the keys line up we care not.
func (m CongressTagMerger) MergeRep(body string, info RepInfo) string {
	return matcher.ReplaceAllStringFunc(body, func(match string) string {
		fieldName, ok := tags[cleanTag(match)]
		if !ok {
			return UnknownTagReplacement
		}

		// This is a fairly ugly use of reflection, but makes it
		// simple to add new tags to the merger. The map contains
		// the field name we want to grab the value from. We switch
		// on types just in case we want to support different types
		// in future releases.
		v := reflect.ValueOf(&info).Elem().FieldByName(fieldName)
		if v.IsValid() {
			switch v.Interface().(type) {
			case string:
				return v.String()
			default:
				// We have no idea how to replace this. Error.
				return UnknownDataTypeReplacement
			}
		} else {
			return InvalidFieldNameReplacement
		}
	})
}

// MergeUnknown will simply replace all tags this merger understands with
// default words.
func (m CongressTagMerger) MergeUnknown(body string) string {
	return matcher.ReplaceAllStringFunc(body, func(match string) string {
		_, ok := tags[cleanTag(match)]
		if !ok {
			return UnknownTagReplacement
		}
		return match
	})
}

// Utility function for cleaning whitespace in tags. Makes matching
// more uniform.
func cleanTag(toClean string) string {
	// Trip all whitespace quickly
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, toClean)
}
