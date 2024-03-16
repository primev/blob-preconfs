package auction_test

import (
	"blob-preconfs/pkg/auction"
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

type mockRegistry struct {
	isRegisteredCallback func(address common.Address) bool
}

func (r *mockRegistry) IsRegisteredOnSettlementLayer(address common.Address) bool {
	return r.isRegisteredCallback(address)
}

func TestAuctionEndsAfterDuration(t *testing.T) {
	periods := []time.Duration{
		100 * time.Millisecond,
		1 * time.Second,
		10 * time.Second,
	}
	for _, period := range periods {
		t.Run(fmt.Sprintf("bidding period %v", period), func(t *testing.T) {
			auction := auction.NewRelayAuction(slog.Default(), &mockRegistry{})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			auctionResultChan := auction.StartAsync(ctx, period)
			fmt.Println(auctionResultChan)

			select {
			case bid := <-auctionResultChan:
				assert.Nil(t, bid, "Auction should end without any bids")
			case <-time.After(period + 10*time.Millisecond):
				t.Fatalf("Auction did not end within the expected time")
			}
		})
	}
}

func TestInvalidBidsIgnored(t *testing.T) {
	relayAuction := auction.NewRelayAuction(slog.Default(), &mockRegistry{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auctionResultChan := relayAuction.StartAsync(ctx, 5*time.Second)

	// Submit invalid (non-whitelisted) bids while auction is running
	for i := 0; i < 20; i++ {
		pk, _ := crypto.GenerateKey()
		bid := auction.MustCreateSignedBid(big.NewInt(100), big.NewInt(100), pk)
		relayAuction.SubmitBid(*bid)
	}

	select {
	case bid := <-auctionResultChan:
		assert.Nil(t, bid, "Auction should end without any bids")
	case <-time.After(6 * time.Second):
		assert.Fail(t, "Auction did not end within the expected time")
	}
}

func TestValidBids(t *testing.T) {
	mockRegistry := &mockRegistry{
		isRegisteredCallback: func(address common.Address) bool {
			return true
		},
	}
	relayAuction := auction.NewRelayAuction(slog.Default(), mockRegistry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auctionResultChan := relayAuction.StartAsync(ctx, 5*time.Second)

	pk1, _ := crypto.HexToECDSA("27ba389e95214192690a05d46716c5e8a1a91922441f29da3bdfbf5c57bcb494")
	pk2, _ := crypto.HexToECDSA("1a51d1c8b33281390cc59928fde876d0577fce196cb66edcf944c4e6b875e980")

	bid1 := auction.MustCreateSignedBid(big.NewInt(100), big.NewInt(999), pk1)
	relayAuction.SubmitBid(*bid1)
	time.Sleep(1 * time.Second)
	currentBid := relayAuction.GetCurrentBid()
	assert.EqualValues(t, bid1.AmountWei, currentBid.AmountWei, "The current bid should match the last submitted bid")
	assert.EqualValues(t, bid1.L1Block, currentBid.L1Block, "The current bid should match the last submitted bid")
	assert.EqualValues(t, bid1.Address, currentBid.Address, "The current bid should match the last submitted bid")
	assert.EqualValues(t, bid1.Signature, currentBid.Signature, "The current bid should match the last submitted bid")

	bid2 := auction.MustCreateSignedBid(big.NewInt(114), big.NewInt(999), pk2)
	relayAuction.SubmitBid(*bid2)
	time.Sleep(1 * time.Second)
	currentBid = relayAuction.GetCurrentBid()
	assert.EqualValues(t, bid2.AmountWei, currentBid.AmountWei, "The current bid should match the last submitted bid")
	assert.EqualValues(t, bid2.L1Block, currentBid.L1Block, "The current bid should match the last submitted bid")
	assert.EqualValues(t, bid2.Address, currentBid.Address, "The current bid should match the last submitted bid")
	assert.EqualValues(t, bid2.Signature, currentBid.Signature, "The current bid should match the last submitted bid")

	select {
	case bid := <-auctionResultChan:
		assert.NotZero(t, bid, "Auction should end with a bid")
		assert.EqualValues(t, bid2.AmountWei, bid.AmountWei, "The winning bid should match the last submitted bid")
		assert.EqualValues(t, bid2.L1Block, bid.L1Block, "The winning bid should match the last submitted bid")
		assert.EqualValues(t, bid2.Address, bid.Address, "The winning bid should match the last submitted bid")
		assert.EqualValues(t, bid2.Signature, bid.Signature, "The winning bid should match the last submitted bid")
	case <-time.After(6 * time.Second):
		assert.Fail(t, "Auction did not end within the expected time")
	}
}
