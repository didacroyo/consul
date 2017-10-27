package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/serf/coordinate"
)

func TestCoordinate_Datacenters(t *testing.T) {
	t.Parallel()
	a := NewTestAgent(t.Name(), "")
	defer a.Shutdown()

	req, _ := http.NewRequest("GET", "/v1/coordinate/datacenters", nil)
	resp := httptest.NewRecorder()
	obj, err := a.srv.CoordinateDatacenters(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	maps := obj.([]structs.DatacenterMap)
	if len(maps) != 1 ||
		maps[0].Datacenter != "dc1" ||
		len(maps[0].Coordinates) != 1 ||
		maps[0].Coordinates[0].Node != a.Config.NodeName {
		t.Fatalf("bad: %v", maps)
	}
}

func TestCoordinate_Nodes(t *testing.T) {
	t.Parallel()
	a := NewTestAgent(t.Name(), "")
	defer a.Shutdown()

	// Make sure an empty list is non-nil.
	req, _ := http.NewRequest("GET", "/v1/coordinate/nodes?dc=dc1", nil)
	resp := httptest.NewRecorder()
	obj, err := a.srv.CoordinateNodes(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates := obj.(structs.Coordinates)
	if coordinates == nil || len(coordinates) != 0 {
		t.Fatalf("bad: %v", coordinates)
	}

	// Register the nodes.
	nodes := []string{"foo", "bar"}
	for _, node := range nodes {
		req := structs.RegisterRequest{
			Datacenter: "dc1",
			Node:       node,
			Address:    "127.0.0.1",
		}
		var reply struct{}
		if err := a.RPC("Catalog.Register", &req, &reply); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Send some coordinates for a few nodes, waiting a little while for the
	// batch update to run.
	arg1 := structs.CoordinateUpdateRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Segment:    "alpha",
		Coord:      coordinate.NewCoordinate(coordinate.DefaultConfig()),
	}
	var out struct{}
	if err := a.RPC("Coordinate.Update", &arg1, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg2 := structs.CoordinateUpdateRequest{
		Datacenter: "dc1",
		Node:       "bar",
		Coord:      coordinate.NewCoordinate(coordinate.DefaultConfig()),
	}
	if err := a.RPC("Coordinate.Update", &arg2, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Query back and check the nodes are present and sorted correctly.
	req, _ = http.NewRequest("GET", "/v1/coordinate/nodes?dc=dc1", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNodes(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 2 ||
		coordinates[0].Node != "bar" ||
		coordinates[1].Node != "foo" {
		t.Fatalf("bad: %v", coordinates)
	}

	// Filter on a nonexistant node segment
	req, _ = http.NewRequest("GET", "/v1/coordinate/nodes?segment=nope", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNodes(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 0 {
		t.Fatalf("bad: %v", coordinates)
	}

	// Filter on a real node segment
	req, _ = http.NewRequest("GET", "/v1/coordinate/nodes?segment=alpha", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNodes(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 1 || coordinates[0].Node != "foo" {
		t.Fatalf("bad: %v", coordinates)
	}

	// Make sure the empty filter works
	req, _ = http.NewRequest("GET", "/v1/coordinate/nodes?segment=", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNodes(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 1 || coordinates[0].Node != "bar" {
		t.Fatalf("bad: %v", coordinates)
	}
}

func TestCoordinate_Node(t *testing.T) {
	t.Parallel()
	a := NewTestAgent(t.Name(), "")
	defer a.Shutdown()

	// Make sure an empty list is non-nil.
	req, _ := http.NewRequest("GET", "/v1/coordinate/node/foo?dc=dc1", nil)
	resp := httptest.NewRecorder()
	obj, err := a.srv.CoordinateNode(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates := obj.(structs.Coordinates)
	if coordinates == nil || len(coordinates) != 0 {
		t.Fatalf("bad: %v", coordinates)
	}

	// Register the nodes.
	nodes := []string{"foo", "bar"}
	for _, node := range nodes {
		req := structs.RegisterRequest{
			Datacenter: "dc1",
			Node:       node,
			Address:    "127.0.0.1",
		}
		var reply struct{}
		if err := a.RPC("Catalog.Register", &req, &reply); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Send some coordinates for a few nodes, waiting a little while for the
	// batch update to run.
	arg1 := structs.CoordinateUpdateRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Segment:    "alpha",
		Coord:      coordinate.NewCoordinate(coordinate.DefaultConfig()),
	}
	var out struct{}
	if err := a.RPC("Coordinate.Update", &arg1, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg2 := structs.CoordinateUpdateRequest{
		Datacenter: "dc1",
		Node:       "bar",
		Coord:      coordinate.NewCoordinate(coordinate.DefaultConfig()),
	}
	if err := a.RPC("Coordinate.Update", &arg2, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Query back and check the nodes are present and sorted correctly.
	req, _ = http.NewRequest("GET", "/v1/coordinate/node/foo?dc=dc1", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNode(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 1 ||
		coordinates[0].Node != "foo" {
		t.Fatalf("bad: %v", coordinates)
	}

	// Filter on a nonexistant node segment
	req, _ = http.NewRequest("GET", "/v1/coordinate/node/foo?segment=nope", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNode(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 0 {
		t.Fatalf("bad: %v", coordinates)
	}

	// Filter on a real node segment
	req, _ = http.NewRequest("GET", "/v1/coordinate/node/foo?segment=alpha", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNode(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 1 || coordinates[0].Node != "foo" {
		t.Fatalf("bad: %v", coordinates)
	}

	// Make sure the empty filter works
	req, _ = http.NewRequest("GET", "/v1/coordinate/node/foo?segment=", nil)
	resp = httptest.NewRecorder()
	obj, err = a.srv.CoordinateNode(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	coordinates = obj.(structs.Coordinates)
	if len(coordinates) != 0 {
		t.Fatalf("bad: %v", coordinates)
	}
}

func TestCoordinate_Update(t *testing.T) {
	t.Parallel()
	a := NewTestAgent(t.Name(), "")
	defer a.Shutdown()

	// Register the node.
	reg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
	}
	var reply struct{}
	if err := a.RPC("Catalog.Register", &reg, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Update the coordinates and wait for it to complete.
	coord := coordinate.NewCoordinate(coordinate.DefaultConfig())
	coord.Height = -5.0
	body := structs.CoordinateUpdateRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Coord:      coord,
	}
	req, _ := http.NewRequest("PUT", "/v1/coordinate/update", jsonReader(body))
	resp := httptest.NewRecorder()
	_, err := a.srv.CoordinateUpdate(resp, req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// Query back and check the coordinates are present.
	args := structs.NodeSpecificRequest{Node: "foo", Datacenter: "dc1"}
	var coords structs.IndexedCoordinates
	if err := a.RPC("Coordinate.Node", &args, &coords); err != nil {
		t.Fatalf("err: %s", err)
	}

	coordinates := coords.Coordinates
	if len(coordinates) != 1 ||
		coordinates[0].Node != "foo" {
		t.Fatalf("bad: %v", coordinates)
	}
}
