package pac

import "testing"

func TestPACEscape(t *testing.T) {
	t.Parallel()
	{
		direct := getDirectList()
		blocked := getBlockedList()
		if direct != "" {
			t.Fail()
		}
		if blocked != "" {
			t.Fail()
		}
	}
	{
		UpdateBlockedList([]string{})
		direct := getDirectList()
		blocked := getBlockedList()
		if direct != "" {
			t.Fail()
		}
		if blocked != "alkasir.com" {
			t.Fail()
		}
	}
	{
		UpdateBlockedList([]string{"test1", "\"buu\"", "test2"})
		UpdateDirectList([]string{"\n\n<>>''"})
		direct := getDirectList()
		blocked := getBlockedList()
		if direct != `\u000A\u000A\x3C\x3E\x3E\'\'` {
			t.Fail()
		}
		if blocked != `alkasir.com",
"test1",
"\"buu\"",
"test2` {
			t.Fail()
		}
	}
}
