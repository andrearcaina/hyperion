package store

import (
	"encoding/json"
	"errors"
	"fmt"
)

type memberAddresses struct {
	HTTP string `json:"http"`
	GRPC string `json:"grpc"`
}

func (s *Store) resolveMemberAddresses(raftAddress, httpAddress, grpcAddress string) (memberAddresses, error) {
	httpAddress, err := memberAddress(raftAddress, httpAddress, s.node.cfg.HTTPAddress)
	if err != nil {
		return memberAddresses{}, fmt.Errorf("resolve member HTTP address: %w", err)
	}

	grpcAddress, err = memberAddress(raftAddress, grpcAddress, s.node.cfg.GRPCAddress)
	if err != nil {
		return memberAddresses{}, fmt.Errorf("resolve member gRPC address: %w", err)
	}

	return memberAddresses{HTTP: httpAddress, GRPC: grpcAddress}, nil
}

func (s *Store) saveMemberAddresses(nodeID string, addresses memberAddresses) error {
	return s.applyCommand(writeCommand{
		Op:          commandSaveMemberAddresses,
		NodeID:      nodeID,
		HTTPAddress: addresses.HTTP,
		GRPCAddress: addresses.GRPC,
	})
}

func (s *Store) withLeaderAddresses(err error) error {
	var notLeader *NotLeaderError
	if !errors.As(err, &notLeader) || notLeader.LeaderID == "" {
		return err
	}

	decorated := *notLeader
	data, getErr := s.db.Get([]byte(memberKeyPrefix + notLeader.LeaderID))
	if getErr != nil {
		decorated.LeaderHTTPAddress, _ = memberAddress(notLeader.LeaderAddress, "", s.node.cfg.HTTPAddress)
		decorated.LeaderGRPCAddress, _ = memberAddress(notLeader.LeaderAddress, "", s.node.cfg.GRPCAddress)
		return &decorated
	}

	var addresses memberAddresses
	if json.Unmarshal(data, &addresses) == nil {
		decorated.LeaderHTTPAddress = addresses.HTTP
		decorated.LeaderGRPCAddress = addresses.GRPC
	}

	return &decorated
}
