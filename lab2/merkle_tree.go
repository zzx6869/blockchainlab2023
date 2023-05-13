package main

import (
	"bytes"
	"crypto/sha256"
	"math"
)

// MerkleTree represent a Merkle tree
type MerkleTree struct {
	RootNode *MerkleNode
	Leaf     [][]byte
}

// MerkleNode represent a Merkle tree node
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
// implement
func NewMerkleTree(data [][]byte) *MerkleTree {
	l := len(data)
	if l%2 == 1 {
		data = append(data, data[len(data)-1])
	}
	var nodePool []*MerkleNode
	for _, tx := range data {
		nodePool = append(nodePool, NewMerkleNode(nil, nil, tx))
	}
	for len(nodePool) > 1 {
		var tmpNodePool []*MerkleNode
		poollen := len(nodePool)
		if poollen%2 != 0 {
			tmpNodePool = append(tmpNodePool, nodePool[poollen-1])
		}
		for i := 0; i < poollen/2; i++ {
			tmpNodePool = append(tmpNodePool, NewMerkleNode(nodePool[2*i], nodePool[2*i+1], nil))
		}
		nodePool = tmpNodePool
	}
	return &MerkleTree{nodePool[0], data}
}

// NewMerkleNode creates a new Merkle tree node
// implement
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	//data = append(left.Data, right.Data...)
	//hash := sha256.Sum256(data)
	//return &MerkleNode{left, right, hash[:]}
	if data != nil {
		hash := sha256.Sum256(data)
		return &MerkleNode{left, right, hash[:]}
	} else {
		hash := sha256.Sum256(append(left.Data, right.Data...))
		return &MerkleNode{left, right, hash[:]}
	}
}

func (t *MerkleTree) SPVproof(index int) ([][]byte, error) {
	n := len(t.Leaf)
	if index >= n {
		return nil, nil
	}
	left := 0
	right := int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
	var path [][]byte
	node := t.RootNode
	for right-left >= 2 {
		if index < (left+right)/2 {
			path = append(path, node.Right.Data)
			node = node.Left
			right = (left + right) / 2
		} else {
			path = append(path, node.Left.Data)
			node = node.Right
			left = (left + right) / 2
		}
	}
	return path, nil
}

func (t *MerkleTree) VerifyProof(index int, path [][]byte) (bool, error) {
	if index >= len(t.Leaf) {
		return false, nil
	}
	hash := sha256.Sum256(t.Leaf[index])
	for i := len(path) - 1; i >= 0; i-- {
		if index%2 == 1 {
			hash = sha256.Sum256(append(path[i], hash[:]...))
		} else {
			hash = sha256.Sum256(append(hash[:], path[i]...))
		}
		index /= 2
	}
	return bytes.Equal(hash[:], t.RootNode.Data), nil
}
