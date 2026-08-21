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

	decorated := *notLeader // don't mutate the original error

	var addresses memberAddresses
	data, err := s.db.Get([]byte(memberKeyPrefix + notLeader.LeaderID))
	if err == nil {
		_ = json.Unmarshal(data, &addresses)
	}

	// if we don't have the addresses in the store, try to resolve them from the leader's raft address
	if addresses.HTTP == "" {
		addresses.HTTP, _ = memberAddress(notLeader.LeaderAddress, "", s.node.cfg.HTTPAddress)
	}
	if addresses.GRPC == "" {
		addresses.GRPC, _ = memberAddress(notLeader.LeaderAddress, "", s.node.cfg.GRPCAddress)
	}

	decorated.LeaderHTTPAddress = addresses.HTTP
	decorated.LeaderGRPCAddress = addresses.GRPC

	return &decorated
}
