package main

type version struct {
	Version    int    // 版本号
	BestHeight int    // 区块链高度
	AddrFrom   string // 发送者的地址 当一个节点接收到 version 消息，它会检查本节点的区块链是否比 BestHeight 的值更大。如果不是，节点就会请求并下载缺失的块
}
