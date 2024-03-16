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
	currentBid        SignedBid
	currentBidMutex   sync.RWMutex // Protects access to currentBid
	auctionResultChan chan SignedBid
	relayRegistry     RelayRegistry
}

func NewRelayAuction(logger *slog.Logger, relayRegistry RelayRegistry) *RelayAuction {
	return &RelayAuction{
		logger:            logger,
		bidSubmissionChan: make(chan SignedBid, 10),
		currentBid:        SignedBid{},
		auctionResultChan: make(chan SignedBid),
		relayRegistry:     relayRegistry,
	}
}

type RelayRegistry interface {
	IsRegisteredOnSettlementLayer(address common.Address) bool
}

func (r *RelayAuction) StartAsync(ctx context.Context, biddingPeriod time.Duration) chan SignedBid {
	go r.runAuction(ctx, biddingPeriod)
	return r.auctionResultChan
}

func (r *RelayAuction) SubmitBid(signedBid SignedBid) {
	r.bidSubmissionChan <- signedBid
}

func (r *RelayAuction) GetCurrentBid() SignedBid {
	r.currentBidMutex.RLock()
	defer r.currentBidMutex.RUnlock()
	return r.currentBid
}

func (r *RelayAuction) runAuction(ctx context.Context, biddingPeriod time.Duration) {
	r.logger.Info("starting auction")
	auctionTimer := time.NewTimer(biddingPeriod)

	for {
		select {
		case <-ctx.Done():
			return
		case <-auctionTimer.C:
			r.currentBidMutex.RLock()
			winner := r.currentBid
			r.currentBidMutex.RUnlock()

			r.logger.Info("auction ended, winner", "bid", winner)
			r.auctionResultChan <- winner
			return
		case bid := <-r.bidSubmissionChan:
			r.logger.Info("new bid received, it will be evaluated", "bid", bid)
			if r.evaluateBid(bid) {
				r.currentBidMutex.Lock()
				r.currentBid = bid
				r.currentBidMutex.Unlock()
			}
		}
	}
}

var relayWhitelist = []common.Address{
	common.HexToAddress("0xDeFEA225C9e43F1A4Ccb561867Be9c9bf3142a98"),
	common.HexToAddress("0xE882aFBf387B7C487b3C17159ad46E13474D9e1E"),
}

func (r *RelayAuction) evaluateBid(bid SignedBid) bool {
	if !bid.Verify() {
		r.logger.Warn("invalid bid received", "bid", bid)
		return false
	}

	if !contains(relayWhitelist, bid.Address) {
		r.logger.Warn("bidder not on whitelist", "bid", bid)
		return false
	}

	if !r.relayRegistry.IsRegisteredOnSettlementLayer(bid.Address) {
		r.logger.Warn("bidder not registered or prepaid on settlement layer", "bid", bid)
		return false
	}

	r.currentBidMutex.RLock()
	isFirstOrHigherBid := SignedBid{}.Address == r.currentBid.Address || bid.AmountWei.Cmp(r.currentBid.AmountWei) > 0
	r.currentBidMutex.RUnlock()
	if isFirstOrHigherBid {
		r.logger.Info("higher or first valid bid received", "bid", bid)
		return true
	}

	return false
}

func contains(slice []common.Address, item common.Address) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
