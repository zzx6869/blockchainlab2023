#  Blockchain lab1
> Name: 张展翔
> Student Number: PB20111669
## 实验题目 
椭圆曲线算法
## 实验目的 
- 理解非对称加密算法
- 理解椭圆曲线算法ECC
- 实现比特币上的椭圆曲线secp256k1算法
## 实验环境
Windows11
Go 1.20.3
GoLand 2023.1
## 实验内容
**secp256k1基本流程**
### 签名流程
1. 我们已知z和满足eG=P的e。
2. 随机选取k。
3. 计算R=kG,及其x轴坐标r。
4. 计算 s=(z+re)/k。
5. (r,s) 即为签名结果。

由补充部分`取1/s的操作通过费马小定理来实现，在函数中Inv(s *big.Int, N *big.Int) *big.Int 已经实现。`
具体代码及注释如下
```go
func (ecc *MyECC) Sign(msg []byte, secKey *big.Int) (*Signature, error) {
	k, err := newRand() //使用newRand函数获取随机k
	if err != nil {
		return nil, err
	}
	R := Multi(G, k) //计算R=kG
	r := R.X//计算R的x轴坐标r
	invk := Inv(k, N)//计算k的逆元
	z := new(big.Int).SetBytes(crypto.Keccak256(msg))
	re := new(big.Int).Mul(r, secKey)
	sum := new(big.Int).Add(z, re)
	s := new(big.Int).Mul(sum, invk)//计算s
	return &Signature{s, r}, nil//返回签名结果(r,s)
}
```
### 验证流程
1. 接收签名者提供的(r,s)作为签名，z是被签名的内容的哈希值。P是签名者的公钥（或者公开的点）。
2. 计算 u=z/s 和 v=r/s。
3. 计算 uG + vP = R。
4. 如果R的x轴坐标等于r，则签名是有效的

且**群的阶数为N，在对于数据操作之后，需要进行求余运算MOD**
具体代码及注释如下:
```go
func (ecc *MyECC) VerifySignature(msg []byte, signature *Signature, pubkey *Point) bool {
    z := new(big.Int).SetBytes(crypto.Keccak256(msg))
    u := new(big.Int).Mul(z, Inv(signature.s, N))//计算u并取MOD
    u.Mod(u, N)
    v := new(big.Int).Mul(signature.r, Inv(signature.s, N))//计算v并取MOD
    v.Mod(v, N)
    uG := Multi(G, u)
    vP := Multi(pubkey, v)
    R := Add(uG, vP)//求R=uG+vP
    return R.X.Cmp(signature.r) == 0//将R的x轴坐标与r比较判断签名是否有效
}
```
## 实验结果
`verify true` 
`verify false`
## 实验总结及建议
难度适中，主要难点在于文档的理解和GO语言的学习上，建议嘛就是网站实在不太方便，容器还总是卡，不如本地开发环境，后续可以优化一下