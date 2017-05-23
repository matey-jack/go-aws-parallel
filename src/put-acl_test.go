package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func strings(a ...string) []string{
	return a
}

func TestGrants(t *testing.T) {
	g := Grants{
		Read: strings("test"),
	}
	i := createPutAclInput("bucket", "pref/key", g)
	assert.Equal(t, "id=test", *i.GrantRead)
}

func TestConcat(t *testing.T) {
	assert.Equal(t, "id=test, id=best", idList(strings("test", "best")))
}
