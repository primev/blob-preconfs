# Auction Package

`auction` contains a simple open auction implementation for relays to bid for rights to be the blob preconfer of a block. This serves as an off-chain method of implementing the relay auction, as opposed to the auction protocol living as contracts on the settlement layer.

Note this auction implementation should be integrated with the oracle service, and a simple contract on the settlement layer. Since bids are managed by the oracle, and the bidding process does not involve sending ether on the mev-commit chain, the auction module needs to confirm that bidding relays have staked/prepaid enough ether in an appropriate contract on the sl. This will be implemented via the `IsRegisteredOnSettlementLayer` hook in `auction.go`, depending on contract implementation.

Following a finished auction, the oracle account will submit a permissioned tx to the settlement layer to finalize the auction winner, which processes the winning relay's prepaid bid. Finally, the oracle will monitor L1 for reward/slashing settlement logic.
