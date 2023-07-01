package main

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 8

func (pow *ProofOfWork) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			IntToHex(pow.block.Header.Version),
			pow.block.Header.PrevBlockHash[:],
			pow.block.Header.MerkleRoot[:],
			IntToHex(pow.block.Header.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(pow.block.Header.Nonce),
		},
		[]byte{},
	)

	return data
}

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

// Run performs a proof-of-work
// implement
func (pow *ProofOfWork) Run() (int64, []byte) {
	nonce := int64(0)
	var hash [32]byte
	pow.block.SetNonce(nonce)
	// ⽐较 hash 和 target
	for nonce < int64(maxNonce) {
		data := pow.prepareData()
		hash = sha256.Sum256(data)
		if new(big.Int).SetBytes(hash[:]).Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
			pow.block.SetNonce(nonce)
		}
	}

	return nonce, nil
}

// Validate validates block's PoW
// implement
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData()
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}
