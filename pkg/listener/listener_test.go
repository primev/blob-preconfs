package listener_test

import (
	"context"
	"log/slog"
	"math/big"
	"sync"
	"testing"
	"time"

	"blob-preconfs/pkg/auction"
	"blob-preconfs/pkg/listener"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

type MockEthClient struct {
	mu             sync.Mutex
	blockNumber    uint64
	incrementEvery time.Duration
}

func NewMockEthClient(initialBlockNumber uint64) *MockEthClient {
	client := &MockEthClient{
		blockNumber:    initialBlockNumber,
		incrementEvery: 12 * time.Second,
	}
	go client.startIncrementing()
	return client
}

func (m *MockEthClient) startIncrementing() {
	ticker := time.NewTicker(m.incrementEvery)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		m.blockNumber++
		m.mu.Unlock()
	}
}

func (m *MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.blockNumber, nil
}

type mockRelayRegistry struct{}

func (r *mockRelayRegistry) IsRegisteredOnSettlementLayer(relayAddress common.Address) bool {
	return true
}

func TestListenerSimulation(t *testing.T) {
	logger := slog.Default()
	mockEthClient := NewMockEthClient(100)
	mockRelayRegistry := &mockRelayRegistry{}

	l := listener.NewListener(logger, mockEthClient, mockRelayRegistry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, auctionWonChan, err := l.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	pk1, _ := crypto.HexToECDSA("27ba389e95214192690a05d46716c5e8a1a91922441f29da3bdfbf5c57bcb494")
	time.Sleep(5 * time.Second)
	blockNum := l.MustGetBlockNum()
	signedBid := auction.MustCreateSignedBid(big.NewInt(43), big.NewInt(int64(blockNum)), pk1)
	require.True(t, signedBid.Verify())
	require.NoError(t, l.SubmitBid(*signedBid))

	testTimeout := time.After(20 * time.Second)

	select {
	case bid := <-auctionWonChan:
		t.Logf("Auction won: %+v", bid)
		require.EqualValues(t, *signedBid, bid)
		return
	case <-testTimeout:
		t.Fatal("Test timed out waiting for auction win")
	case <-ctx.Done():
		t.Error("Context was canceled before auction win could be received")
	}
}
