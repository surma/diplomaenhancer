package hostfile

import (
	"testing"
	"io/ioutil"
	"os"
)

func TestParser(t *testing.T) {
	testfile := "testdata/hosts1"
	in, e := os.Open(testfile+".txt")
	if e != nil {
		t.Fatalf("Could not open testfile %s: %s", testfile, e)
	}
	defer in.Close()

	resultf, e := os.Open(testfile+"_parsed.txt")
	if e != nil {
		t.Fatalf("Could not open testfile %s: %s", testfile, e)
	}
	result, e := ioutil.ReadAll(resultf)
	if e != nil {
		t.Fatalf("Could not read testfile %s: %s", testfile, e)
	}
	resultf.Close()

	h, e := Parse(in)
	if e != nil {
		t.Fatalf("Could not parse testfile %s: %s", testfile, e)
	}
	if string(result) != h.String() {
		t.Logf("Result: \"%s\"", h.String())
		t.Fatalf("Did not parse correctly")
	}
}


func TestFailure(t *testing.T) {
	testfile := "testdata/hosts2"
	in, e := os.Open(testfile+".txt")
	if e != nil {
		t.Fatalf("Could not open testfile %s: %s", testfile, e)
	}
	defer in.Close()

	_, e = Parse(in)
	if e == nil {
		t.Fatalf("Did not fail! (Which is wrong)")
	}
}
