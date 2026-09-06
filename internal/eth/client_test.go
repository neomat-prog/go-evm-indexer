package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neomat-prog/go-evm-indexer/internal/config"
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

	client, chainID, err := NewClient(context.Background(), config.ConfigOpts{ETHRPCURL: node.URL})

	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if chainID.Int64() != 1 {
		t.Fatalf("chainID = %d, want 1", chainID.Int64())
	}

	defer client.Close()
}

func TestNewClientFailsWhenNodeIsDown(t *testing.T) {
	node := fakeNode(t, "0x1")
	url := node.URL
	node.Close()

	if _, _, err := NewClient(context.Background(), config.ConfigOpts{
		ETHRPCURL: url,
	}); err == nil {
		t.Fatal("NewClient succeeded against a closed node, want error")
	}
}
