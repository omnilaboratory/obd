package service

import (
	"sync"
)

type channelManager struct {
	mu sync.Mutex
}

var ChannelManager channelManager

func (this *channelManager) addChannel() {

}
