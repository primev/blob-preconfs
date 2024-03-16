# Listener Package

This package contains a listener worker, that monitors L1 for new blocks, and starts a new relay auction each time. This module also facilities bid submission and querying. The exported `AuctionWonChan` channel will be useful to subscribe to, so that other oracle workers can post the auction winner to the settlement layer, and follow through with rewards/slashing.
