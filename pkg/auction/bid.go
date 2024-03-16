package auction

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type SignedBid struct {
	AmountWei *big.Int       `json:"amountWei"`
	L1Block   *big.Int       `json:"l1Block"`
	Address   common.Address `json:"address"`
	Signature hexutil.Bytes  `json:"signature"`
}

func CreateSignedBid(amountWei *big.Int, l1Block *big.Int, privateKey *ecdsa.PrivateKey) (*SignedBid, error) {
	hash := getDataHash(amountWei, l1Block)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}
	address, err := getAddressFromSig(signature, hash)
	if err != nil {
		return nil, err
	}
	return &SignedBid{
		AmountWei: amountWei,
		L1Block:   l1Block,
		Address:   address,
		Signature: signature,
	}, nil
}

func (b *SignedBid) Verify() bool {
	hash := getDataHash(b.AmountWei, b.L1Block)
	sigPublicKey, err := crypto.SigToPub(hash.Bytes(), b.Signature)
	if err != nil {
		return false
	}
	signerAddress := crypto.PubkeyToAddress(*sigPublicKey)
	return signerAddress == b.Address
}

func getDataHash(amountWei *big.Int, l1Block *big.Int) common.Hash {
	data := fmt.Sprintf("%s%s", amountWei.String(), l1Block.String())
	return crypto.Keccak256Hash([]byte(data))
}

func getAddressFromSig(signature hexutil.Bytes, hash common.Hash) (common.Address, error) {
	sigPublicKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*sigPublicKey), nil
}
