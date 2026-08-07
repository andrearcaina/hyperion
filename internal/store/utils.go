package store

import (
	"net"
	"strings"
)

func invalidUserKey(key string) bool {
	return key == "" || strings.HasPrefix(key, memberKeyPrefix)
}

func isInternalKey(key []byte) bool {
	return strings.HasPrefix(string(key), memberKeyPrefix)
}

func memberAddress(raftAddress, advertisedAddress, localAddress string) (string, error) {
	if advertisedAddress != "" {
		if _, _, err := net.SplitHostPort(advertisedAddress); err != nil {
			return "", err
		}
		return advertisedAddress, nil
	}

	raftHost, _, err := net.SplitHostPort(raftAddress)
	if err != nil {
		return "", err
	}

	_, localPort, err := net.SplitHostPort(localAddress)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(raftHost, localPort), nil
}
