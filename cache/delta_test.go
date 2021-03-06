package cache

import (
	"imposm3/element"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"testing"
)

func mknode(id int64) element.Node {
	return element.Node{
		OSMElem: element.OSMElem{
			Id: id,
		},
		Long: 8,
		Lat:  10,
	}
}

func TestRemoveSkippedNodes(t *testing.T) {
	nodes := []element.Node{
		mknode(0),
		mknode(1),
		mknode(-1),
		mknode(2),
		mknode(-1),
	}
	nodes = removeSkippedNodes(nodes)
	if l := len(nodes); l != 3 {
		t.Fatal(nodes)
	}
	if nodes[0].Id != 0 || nodes[1].Id != 1 || nodes[2].Id != 2 {
		t.Fatal(nodes)
	}

	nodes = []element.Node{
		mknode(-1),
		mknode(-1),
	}
	nodes = removeSkippedNodes(nodes)
	if l := len(nodes); l != 0 {
		t.Fatal(nodes)
	}

	nodes = []element.Node{
		mknode(-1),
		mknode(1),
		mknode(-1),
		mknode(-1),
		mknode(-1),
		mknode(2),
	}
	nodes = removeSkippedNodes(nodes)
	if l := len(nodes); l != 2 {
		t.Fatal(nodes)
	}
	if nodes[0].Id != 1 || nodes[1].Id != 2 {
		t.Fatal(nodes)
	}
}

func TestReadWriteDeltaCoords(t *testing.T) {
	checkReadWriteDeltaCoords(t, false)
}

func TestReadWriteDeltaCoordsLinearImport(t *testing.T) {
	checkReadWriteDeltaCoords(t, true)
}

func checkReadWriteDeltaCoords(t *testing.T, withLinearImport bool) {
	cache_dir, _ := ioutil.TempDir("", "imposm3_test")
	defer os.RemoveAll(cache_dir)

	cache, err := newDeltaCoordsCache(cache_dir)
	if err != nil {
		t.Fatal()
	}

	if withLinearImport {
		cache.SetLinearImport(true)
	}

	// create list with nodes from Id 0->999 in random order
	nodeIds := rand.Perm(1000)
	nodes := make([]element.Node, 1000)
	for i := 0; i < len(nodes); i++ {
		nodes[i] = mknode(int64(nodeIds[i]))
	}

	// add nodes in batches of ten
	for i := 0; i <= len(nodes)-10; i = i + 10 {
		// sort each batch as required by PutCoords
		sort.Sort(byId(nodes[i : i+10]))
		cache.PutCoords(nodes[i : i+10])
	}

	if withLinearImport {
		cache.SetLinearImport(false)
	}

	for i := 0; i < len(nodes); i++ {
		data, err := cache.GetCoord(int64(i))
		if err == NotFound {
			t.Fatal("missing coord:", i)
		} else if err != nil {
			t.Fatal(err)
		}
		if data.Id != int64(i) {
			t.Errorf("unexpected result of GetNode: %v", data)
		}
	}

	_, err = cache.GetCoord(999999)
	if err != NotFound {
		t.Error("missing node returned not NotFound")
	}

	// test delete
	cache.PutCoords([]element.Node{mknode(999999)})

	_, err = cache.GetCoord(999999)
	if err == NotFound {
		t.Error("missing coord")
	}
	err = cache.DeleteCoord(999999)
	if err != nil {
		t.Fatal(err)
	}
}
