package auction_test

import (
	"blob-preconfs/pkg/auction"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	expectedAddr  = common.HexToAddress("0xDeFEA225C9e43F1A4Ccb561867Be9c9bf3142a98")
	privateKey, _ = crypto.HexToECDSA("27ba389e95214192690a05d46716c5e8a1a91922441f29da3bdfbf5c57bcb494")
)

func TestCreateSignedBid(t *testing.T) {
	derivedAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	assert.Equal(t, expectedAddr, derivedAddr)

	signedBid, err := auction.CreateSignedBid(big.NewInt(677), big.NewInt(1234567), privateKey)

	assert.NoError(t, err)
	assert.NotNil(t, signedBid)
	assert.Equal(t, big.NewInt(677), signedBid.AmountWei)
	assert.Equal(t, big.NewInt(1234567), signedBid.L1Block)
	assert.NotEmpty(t, signedBid.Signature)
	assert.Equal(t, expectedAddr, signedBid.Address)
}

func TestVerifySignedBid(t *testing.T) {
	amount := big.NewInt(677)
	l1Block := big.NewInt(1234567)
	sig := "0x654f553afe2f8eca87582a23817e40a3cdff28e07995136503a999bc5d18b8f62857d603662355e6bc4963cccd426298d771c065d8d9aa880345dc88a4e791ce00"

	signedBid := &auction.SignedBid{
		AmountWei: amount,
		L1Block:   l1Block,
		Address:   expectedAddr,
		Signature: hexutil.MustDecode(sig),
	}
	assert.True(t, signedBid.Verify())
}

func TestEncodeSignedBid(t *testing.T) {
	signedBid, err := auction.CreateSignedBid(big.NewInt(677), big.NewInt(1234567), privateKey)
	assert.NoError(t, err)

	encoded := auction.EncodeSignedBid(signedBid)
	expectedJson := `{"amountWei":677,"l1Block":1234567,"address":"0xdefea225c9e43f1a4ccb561867be9c9bf3142a98","signature":"0x654f553afe2f8eca87582a23817e40a3cdff28e07995136503a999bc5d18b8f62857d603662355e6bc4963cccd426298d771c065d8d9aa880345dc88a4e791ce00"}`
	assert.Equal(t, expectedJson, encoded)
}

func TestDecodeSignedBid(t *testing.T) {
	encoded := `{"amountWei":677,"l1Block":1234567,"address":"0xdefea225c9e43f1a4ccb561867be9c9bf3142a98","signature":"0x654f553afe2f8eca87582a23817e40a3cdff28e07995136503a999bc5d18b8f62857d603662355e6bc4963cccd426298d771c065d8d9aa880345dc88a4e791ce00"}`
	signedBid, err := auction.DecodeSignedBid(encoded)
	assert.NoError(t, err)
	assert.NotNil(t, signedBid)
	assert.Equal(t, big.NewInt(677), signedBid.AmountWei)
	assert.Equal(t, big.NewInt(1234567), signedBid.L1Block)
	assert.Equal(t, expectedAddr, signedBid.Address)
	assert.Equal(t, "0x654f553afe2f8eca87582a23817e40a3cdff28e07995136503a999bc5d18b8f62857d603662355e6bc4963cccd426298d771c065d8d9aa880345dc88a4e791ce00", hexutil.Encode(signedBid.Signature))
	assert.True(t, signedBid.Verify())
}
