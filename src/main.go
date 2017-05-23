package main

import (
	"os"
)

/*
	PRO TIPP: when giving rights to other accounts, always include all the existing
	rights for yourself or otherwise you'll not get back at the data!
 */
func run() int {
	canonicalUserOne := "drop your really long canonical id here"
	canonicalUserTwo := "another one ..."
	g := Grants{
		FullControl: (&[...]string{canonicalUserOne, canonicalUserTwo})[:],
	}
	err := putAclRecursive("some-test-bucket", "test-acls/", g)
	if err==nil {
		return 0
	}
	return 1
}

func main() {
	os.Exit(run())
}

