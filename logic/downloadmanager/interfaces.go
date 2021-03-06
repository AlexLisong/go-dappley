package downloadmanager

import (
	"github.com/dappley/go-dappley/common/pubsub"
	"github.com/dappley/go-dappley/network/networkmodel"
	"github.com/golang/protobuf/proto"
)

type NetService interface {
	GetPeers() []networkmodel.PeerInfo
	GetHostPeerInfo() networkmodel.PeerInfo
	UnicastNormalPriorityCommand(commandName string, message proto.Message, destination networkmodel.PeerInfo)
	UnicastHighProrityCommand(commandName string, message proto.Message, destination networkmodel.PeerInfo)
	BroadcastNormalPriorityCommand(commandName string, message proto.Message)
	BroadcastHighProrityCommand(commandName string, message proto.Message)
	Listen(subscriber pubsub.Subscriber)
}
