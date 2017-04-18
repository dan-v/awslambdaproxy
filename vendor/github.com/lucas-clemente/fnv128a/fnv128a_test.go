package fnv128a_test

import (
	"testing"

	"github.com/lucas-clemente/fnv128a"
)

func TestNullHash(t *testing.T) {
	hash := fnv128a.New()
	h, l := hash.Sum128()
	if h != 0x6c62272e07bb0142 || l != 0x62b821756295c58d {
		t.FailNow()
	}
}

func TestHash(t *testing.T) {
	hash := fnv128a.New()
	_, err := hash.Write([]byte("foobar"))
	h, l := hash.Sum128()
	if err != nil || h != 0x343e1662793c64bf || l != 0x6f0d3597ba446f18 {
		t.FailNow()
	}
}
