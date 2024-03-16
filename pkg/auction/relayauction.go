package auction

import (
	"context"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type RelayAuction struct {
	logger            *slog.Logger
	bidSubmissionChan chan Bid
	currentBid        *Bid
	currentBidMutex   sync.RWMutex
	auctionResultChan chan Bid
}

func NewRelayAuction(logger *slog.Logger) *RelayAuction {
	return &RelayAuction{
		logger:            logger,
		bidSubmissionChan: make(chan Bid, 10),
		auctionResultChan: make(chan Bid),
	}
}

type Bid struct {
	Bidder    common.Address
	AmountWei *big.Int
}

func (r *RelayAuction) StartAsync(ctx context.Context, biddingPeriod time.Duration) chan Bid {
	go r.runAuction(ctx, biddingPeriod)
	return r.auctionResultChan
}

func (r *RelayAuction) SubmitBid(bidder common.Address, amountWei *big.Int) {
	r.bidSubmissionChan <- Bid{Bidder: bidder, AmountWei: amountWei}
}

func (r *RelayAuction) GetCurrentBid() *Bid {
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
			if r.currentBid != nil {
				r.auctionResultChan <- *r.currentBid
			}
			r.currentBidMutex.Unlock()
			close(r.auctionResultChan) // Close the channel to signal completion of the auction
			return
		case bid := <-r.bidSubmissionChan:
			r.currentBidMutex.Lock()
			if r.currentBid == nil || bid.AmountWei.Cmp(r.currentBid.AmountWei) > 0 {
				r.currentBid = &bid
			}
			r.currentBidMutex.Unlock()
		}
	}
}
