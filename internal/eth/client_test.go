package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeNode(t *testing.T, chainID string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%q}`, req.ID, chainID)
	}))
	t.Cleanup(srv.Close)

	return srv
}

func TestNewClientConnects(t *testing.T) {
	node := fakeNode(t, "0x1")

	client, err := NewClient(context.Background(), ConfigOpts{RPCServerAddr: node.URL})

	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()
}

func TestNewClientFailsWhenNodeIsDown(t *testing.T) {
	node := fakeNode(t, "0x1")
	url := node.URL
	node.Close()

	if _, err := NewClient(context.Background(), ConfigOpts{
		RPCServerAddr: url,
	}); err == nil {
		t.Fatal("NewClient succeeded against a closed node, want error")
	}
}
