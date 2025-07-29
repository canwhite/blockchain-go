## Byzantine Fault Tolerance

拜占庭将军问题是分布式系统中达成共识的难题，将军们需统一行动，但可能有叛徒传递错误信息。工作量证明（PoW）通过计算难题解决此问题：一个营地解决 PoW 后，广播结果，其他营地接受并延续此消息链，丢弃竞争消息。若解决营地篡改数据，其他营地可快速识别错误哈希，拒绝其结果，转而接受下一个解决营地的消息。这激励诚实行为，因为篡改者浪费算力无回报，从而保证系统一致性和安全性

> It’s important to note here that Proof of Work does not care about the message itself, only that the nodes agreed to the final message.  
> PoW 的核心：达成共识

## Smart Contracts & Turing Completeness

- Smart Contracts  
  如果说比特币是跨国币的话，可以把 Ethereum 看作跨国合同
- Turing Completeness  
  是说 Solidity 是一门具有完全功能的编程语言，能在 Ethereum Blockchain 上编写合约

## Delegated Proof of Stake (DPOS)

持币者投票选出 21~101 个节点，按顺序轮流出块，3 秒确认，高效节能，支持高 TPS，适合商用链，兼顾去中心化与性能
DPoS 收益分配 = 节点奖励 + 投票者分红 + 生态基金，具体规则由链上治理决定。

## State Channels / Side Chains

- State Channels
- Side Chains
