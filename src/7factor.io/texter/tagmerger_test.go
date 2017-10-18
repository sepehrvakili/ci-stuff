package texter

import (
	"testing"
)

func Test_MergeUnkownWorks(context *testing.T) {
	var table = []struct {
		message  string
		expected string
	}{
		{
			"",
			"",
		}, {
			"garbled garbage is garbage ;:lkjkje {} {{} {",
			"garbled garbage is garbage ;:lkjkje {} {{} {",
		}, {
			"{{ unknown-tag }} {{ unknown-tag-2 }}",
			UnknownTagReplacement + " " + UnknownTagReplacement,
		}, {
			"{{ unknown-tag }} {{ targets.title }}",
			UnknownTagReplacement + " {{ targets.title }}",
		}, {
			"{{ unknown tag }}",
			UnknownTagReplacement,
		},
	}

	merger := NewTagMerger()
	for _, test := range table {
		actual := merger.MergeUnknown(test.message)
		if actual != test.expected {
			context.Errorf("MergeUnknown(%q) => %q, want %q", test.message,
				actual, test.expected)
		}
	}
}

func Test_MergeWorks(context *testing.T) {
	var table = []struct {
		message  string
		rep      RepInfo
		expected string
	}{
		{
			"",
			RepInfo{},
			"",
		},
		{
			"hello world",
			RepInfo{},
			"hello world",
		},
		{
			"{{ unknown-tag }}",
			RepInfo{},
			UnknownTagReplacement,
		},
		{
			"{{ targets.title }}",
			RepInfo{},
			"",
		},
		{
			"{{ targets.TiTle }}",
			RepInfo{
				LongTitle: "Representative",
			},
			"Representative",
		}, {
			"Call {{ targets.title }} {{ targets.full_name }} at {{ targets.phone }}",
			RepInfo{
				LongTitle:    "Representative",
				OfficialName: "Bob Smith",
				PhoneNumber:  "4045551234",
			},
			"Call Representative Bob Smith at 4045551234",
		}, {
			"{{ targets.last_name }}",
			RepInfo{
				LastName: "Smith",
			},
			"Smith",
		}, {
			"{{ targets.short_title }}",
			RepInfo{
				Title: "Rep.",
			},
			"Rep.",
		},
	}

	merger := NewTagMerger()
	for _, test := range table {
		actual := merger.MergeRep(test.message, test.rep)
		if actual != test.expected {
			context.Errorf("MergeRep(%q, %q) => %q, want %q", test.message, test.rep,
				actual, test.expected)
		}
	}
}
