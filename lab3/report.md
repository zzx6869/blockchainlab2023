

# 区块链lab3

> Name: 张展翔
>
> Student Number：PB20111669

## 实验题目

基于比特币区块链的简单搭建（下）

## 实验环境

GoLand 

Golang 1.18.1

## 实验内容

### UTXO池部分

主要思路：

大体思路与FindUTXO类似，只需要将不同的交易添加到txID即可

```go
func (u UTXOSet) FindUnspentOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}
```

### POW部分

需要创建一个数据处理函数prepareDate,将区块头内容添加到[]byte中

```go
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
```

run函数采用枚举法即可

```go
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
```

Validate函数需要将区块头的哈希和目标值比较即可
```go
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData()
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}
```

### Blockchain部分

MineBlock

首先需要判断交易是否合法，从k-v数据库中取出上个区块的信息（“l”），和交易记录写进新的区块中，然后把区块写入数据库即可

```go
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash [32]byte

	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		hash := b.Get([]byte("l"))
		copy(lastHash[:], hash)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.CalCulHash(), newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.CalCulHash())
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.CalCulHash()

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}
```

FindUTXO

参照FindTransaction进行遍历区块链即可

```go
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.GetTransactions() {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}
		if block.GetPrevhash() == [32]byte{} {
			break
		}
	}

	return UTXO
}
```

### Transaction部分

首先需要找到from地址所对应的UTXO，然后传入from对应钱包的公钥哈希

ID域通过SetID方法，Vout通过NewTXOutput函数，把金额分为超出交易和没有超出分别处理即可

```go
func NewUTXOTransaction(from, to []byte, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	_wallet := wallets.GetWallet(from)
	pubKeyHash := HashPublicKey(_wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindUnspentOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TXInput{txID, out, nil, _wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	UTXOSet.Blockchain.SignTransaction(&tx, _wallet.PrivateKey)
	return &tx
}
```

