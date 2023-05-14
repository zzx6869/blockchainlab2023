# 区块链Lab2

## 实验题目

区块链的基本结构

## 实验目的

- 了解区块链上的简单数据结构
- 实现Merkle树的构建
- 初步理解UTXO的使用和验证
- 理解比特币上的交易创建

## 实验环境

Windows11

GoLand 2023.1.1

Go 1.20.3

## 主要代码分析

在`merkle_tree.go`中

建立新默克尔树的函数如下

```go
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
```

由实验介绍中的Merkle树部分可知，==在Merkle树构建过程中，我们从底部开始，对节点进行哈希合并操作，直到节点数量减少为1。对于叶子节点，我们会进行哈希加密（在比特币中采用了双重SHA加密哈希的方式,此前实验中我们使用**单次sha256的方式加密**）。如果结点个数为奇数，那么最后一个节点会把最后一个交易复制一份，来保证数量为偶。==故我们每进行一次合并，都要判断节点个数是否为奇数，若为奇数则需要复制一次最后一个节点，直到合并为只有一个节点。

![image-20230513224932221](https://s2.loli.net/2023/05/13/iDcwpk5s4VYGjRr.png)

而对于默克尔树的结点部分，我们只需判断数据域是否为空即可，然后分开处理

```go
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	if data != nil {
		hash := sha256.Sum256(data)
		return &MerkleNode{left, right, hash[:]}
	} else {
		hash := sha256.Sum256(append(left.Data, right.Data...))
		return &MerkleNode{left, right, hash[:]}
	}
}
```

在实现SPV部分中，我们需要完成路径的搜索和验证

具体方法在第一次作业中涉及，我们需要判断i所在的分支，根据不同分支来添加不同的路径顺序，对于验证部分，我们只需判断路径是否有效即可

代码如下：

```go
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
```



在`transaction.go`中，判断是否为coinbase交易时，只需要按照lab2.md中的Coinbase交易部分实现即可，即coinbase交易中对应的输入中`Txid` 为空，`Vout`对应为-1，并且是一个区块的第一笔交易

故代码如下：

```go
func (t *Transaction) IsCoinBase() bool {
	if len(t.Vin[0].Txid) == 0 && t.Vin[0].Vout == -1 && len(t.Vin) == 1 {
		return true
	} else {
		return false
	}
}
```

在`wallet.go`中，我们需要完成的是获取公钥所对应的地址，由说明文档中可知，地址的计算方法如下

![image-20230514155042145](https://s2.loli.net/2023/05/14/7fCFWkZ2h6xLN5t.png)

公钥哈希值的计算函数已经实现，我们直接调用即可，我们所要做的就是加入版本号然后进行双重SHA256哈希加密，并取前4字节作为校验和，然后把版本号，公钥哈希和校验和组合通过Base58加密即可，实现如下

```go
func (w *Wallet) GetAddress() []byte {
	publicKeyHash := HashPublicKey(w.PublicKey)
	versionedPublicKeyHash := append([]byte{version}, publicKeyHash...)
	checkSum := sha256.Sum256(versionedPublicKeyHash)
	checkSum = sha256.Sum256(checkSum[:])
	finalHash := append(versionedPublicKeyHash, checkSum[:checkSumlen]...)
	address := base58.Encode(finalHash)
	return []byte(address)
}
```

在`TXOutput.go`中，我们需要实现设置锁定脚本PubKeyHash部分，由P2PKH的计算方式及设置锁定脚本和publickeyhash对应相同，可以写出锁定代码如下：

```go
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := base58.Decode(string(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}
```

