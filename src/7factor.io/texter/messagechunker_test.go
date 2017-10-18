package texter

import "testing"

func Test_ChunkerWorksAsExpected(context *testing.T) {
	var table = []struct {
		toChunk  string
		expected []string
	}{
		{
			`Hello this is my test that is under 160 characters.`,
			[]string{
				`Hello this is my test that is under 160 characters.`,
			},
		}, {
			`Hello this is my test that is exactly 160 characters. Isn't it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on`,
			[]string{
				`Hello this is my test that is exactly 160 characters. Isn't it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on`,
			},
		}, {
			`Hello this is my test that is moooore 160 characters. Isn't it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on but now it's more than that.`,
			[]string{
				`Hello this is my test that is moooore 160 characters. Isn't it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on`,
				`but now it's more than that.`,
			},
		}, {
			`01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789`,
			[]string{
				`0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789`,
				`0123456789`,
			},
		}, {
			`Hello this is my test that is exactly 160 characters. Isnt it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on Hello this is my test that is exactly 160 characters. Isnt it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on Hello this is my test that is exactly 160 characters. Isnt it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on `,
			[]string{
				`Hello this is my test that is exactly 160 characters. Isnt it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on`,
				`Hello this is my test that is exactly 160 characters. Isnt it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on`,
				`Hello this is my test that is exactly 160 characters. Isnt it great? It's long with lots of characters. I bet we could go on and on and on and on and on and on`,
			},
		}, {
			`Thank you for becoming a CREDO Rapid Responder. We’ll be texting you with our most urgent and time-sensitive campaigns so that you can fight for your progressive values from your phone. If you want to stop receiving Rapid Responder messages, simply text “STOP” at any time.`,
			[]string{
				`Thank you for becoming a CREDO Rapid Responder. We’ll be texting you with our most urgent and time-sensitive campaigns so that you can fight for your`,
				`progressive values from your phone. If you want to stop receiving Rapid Responder messages, simply text “STOP” at any time.`,
			},
		},
	}

	chunker := NewMessageChunker()
	for _, test := range table {
		actual := chunker.Split(test.toChunk)
		if len(actual) != len(test.expected) {
			context.Errorf("Input is %v", test.toChunk)
			context.Fatalf("Count was wrong, want %v got %v.", len(test.expected), len(actual))
		}
		for index, item := range actual {
			if test.expected[index] != item {
				context.Errorf("Split(%q) => %q, want %q", test.toChunk, actual, test.expected)
			}
		}
	}
}
