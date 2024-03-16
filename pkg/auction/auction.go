package auction

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type RelayAuction struct {
	logger            *slog.Logger
	bidSubmissionChan chan SignedBid
	currentBid        *SignedBid
	currentBidMutex   sync.RWMutex
	auctionResultChan chan *SignedBid
}

func NewRelayAuction(logger *slog.Logger) *RelayAuction {
	return &RelayAuction{
		logger:            logger,
		bidSubmissionChan: make(chan SignedBid, 10),
		auctionResultChan: make(chan *SignedBid),
	}
}

func (r *RelayAuction) StartAsync(ctx context.Context, biddingPeriod time.Duration,
) (auctionResultChan chan *SignedBid) { // Nil bid from this chan means auction ended without any bids
	go r.runAuction(ctx, biddingPeriod)
	return r.auctionResultChan
}

func (r *RelayAuction) SubmitBid(signedBid SignedBid) {
	r.bidSubmissionChan <- signedBid
}

func (r *RelayAuction) GetCurrentBid() *SignedBid {
	r.currentBidMutex.RLock()
	defer r.currentBidMutex.RUnlock()
	if r.currentBid == nil {
		return nil
	}
	bidCopy := *r.currentBid
	return &bidCopy
}

func (r *RelayAuction) runAuction(ctx context.Context, biddingPeriod time.Duration) {
	r.logger.Info("starting auction")
	auctionTimer := time.NewTimer(biddingPeriod)

	for {
		select {
		case <-ctx.Done():
			return
		case <-auctionTimer.C:
			r.currentBidMutex.Lock()
			r.auctionResultChan <- r.currentBid
			r.currentBidMutex.Unlock()
			return
		case bid := <-r.bidSubmissionChan:
			r.currentBidMutex.Lock()
			r.evaluateBid(bid)
			r.currentBidMutex.Unlock()
		}
	}
}

var relayWhitelist = []common.Address{
	common.HexToAddress("0xAddress1"),
	common.HexToAddress("0xAddress2"),
	common.HexToAddress("0xAddress3"),
}

func (r *RelayAuction) evaluateBid(bid SignedBid) {

	if !bid.Verify() {
		r.logger.Warn("invalid bid received", "bid", bid)
		return
	}

	if !contains(relayWhitelist, bid.Address) {
		r.logger.Warn("bidder not on whitelist", "bid", bid)
		return
	}

	// TODO: check if bidder is registered on sl via hook

	if r.currentBid == nil || bid.AmountWei.Cmp(r.currentBid.AmountWei) > 0 {
		r.currentBid = &bid
	}
}

func contains(slice []common.Address, item common.Address) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
