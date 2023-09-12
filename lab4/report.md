# 区块链Lab3

> Name:张展翔
>
> Student Number:PB20111669

## 实验题目

智能合约的开发和部署

## 实验目的

- 理解以太坊上智能合约
- 开发实现简单的智能合约
- 部署以太坊上的智能合约

## 实验内容

实现简单的投票合约

### 投票合约的实现

首先创建一个proposal结构体，包含提案名和得票数，然后声明了一个数组proposals用来存储所有提案。

proposal函数完成了创建一个提案的功能，vote函数则是向给定一个提案的编号并向其投一票，winningProposal则是找出票数最多的提案

```solidity
// SPDX-License-Identifier: MIT
pragma solidity >=0.5.17;

contract Vote {
    struct Proposal{
        string name;
        uint voteCount;
    }
    Proposal[] public proposals;

    function proposal(string memory proposal_name) public{
        proposals.push(Proposal(proposal_name, 0));
    }
    
    function vote(uint i) public{
        proposals[i].voteCount += 1;
    }

    function winningProposal() public view 
        returns (string memory) {
        uint winnerVoteCount = 0;
        uint winnerIndex = 0;
        for (uint i = 0; i < proposals.length; i++){
            if (proposals[i].voteCount > winnerVoteCount){
                winnerVoteCount = proposals[i].voteCount;
                winnerIndex = i;
            }
        }
        return proposals[winnerIndex].name;
    }
}
```

### 投票合约的部署

按照实验文档中的步骤进行编译和部署

![image-20230706222229255](https://s2.loli.net/2023/07/06/MezxfNmgrGHv24o.png)

![image-20230706222339476](https://s2.loli.net/2023/07/06/ypWFH5iBoaOxc2U.png)

![image-20230706222354849](https://s2.loli.net/2023/07/06/kiDMOwHx84gsK2A.png)

创建提案A和提案B

![image-20230706222414389](https://s2.loli.net/2023/07/06/4WwILTtfhyUZKpg.png)

向提案A投一票

![image-20230706222513889](https://s2.loli.net/2023/07/06/HziBXqanFgxGhcJ.png)

查看获胜的提案

![image-20230706222535840](https://s2.loli.net/2023/07/06/W84hTtge6UAdKPp.png)