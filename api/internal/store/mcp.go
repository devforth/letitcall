package store

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
)

func (s *Store) PutMCPAuthorization(key string, authorization model.MCPAuthorization) error {
	return putJSON(s.mcpAuthorizations, []byte(key), authorization)
}

func (s *Store) GetMCPAuthorization(key string, now time.Time) (model.MCPAuthorization, error) {
	var authorization model.MCPAuthorization
	if err := getJSON(s.mcpAuthorizations, []byte(key), &authorization); err != nil {
		return model.MCPAuthorization{}, err
	}
	if !authorization.ExpiresAt.After(now) {
		return model.MCPAuthorization{}, ErrNotFound
	}
	return authorization, nil
}

func (s *Store) ConsumeMCPAuthorization(key string, now time.Time) (model.MCPAuthorization, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	authorization, err := s.GetMCPAuthorization(key, now)
	if err != nil {
		return model.MCPAuthorization{}, err
	}
	return authorization, s.mcpAuthorizations.Delete([]byte(key), nil)
}

func (s *Store) PutMCPConnection(connection model.MCPConnection) error {
	return putJSON(s.mcpConnections, []byte(connection.ID), connection)
}

func (s *Store) GetMCPConnection(id string) (model.MCPConnection, error) {
	var connection model.MCPConnection
	if err := getJSON(s.mcpConnections, []byte(id), &connection); err != nil {
		return model.MCPConnection{}, err
	}
	return connection, nil
}

func (s *Store) ModifyMCPConnection(id string, modify func(*model.MCPConnection) error) (model.MCPConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	connection, err := s.GetMCPConnection(id)
	if err != nil {
		return model.MCPConnection{}, err
	}
	if err := modify(&connection); err != nil {
		return model.MCPConnection{}, err
	}
	return connection, s.PutMCPConnection(connection)
}

func (s *Store) ListMCPConnections(userEmail string) ([]model.MCPConnection, error) {
	iterator := s.mcpConnections.NewIterator(nil, nil)
	defer iterator.Release()
	connections := make([]model.MCPConnection, 0)
	for iterator.Next() {
		var connection model.MCPConnection
		if err := json.Unmarshal(iterator.Value(), &connection); err != nil {
			return nil, err
		}
		if connection.UserEmail == normalizeEmail(userEmail) {
			connections = append(connections, connection)
		}
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(connections, func(i, j int) bool { return connections[i].CreatedAt.Before(connections[j].CreatedAt) })
	return connections, nil
}

func (s *Store) DeleteMCPConnection(id, userEmail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	connection, err := s.GetMCPConnection(id)
	if err != nil {
		return err
	}
	if connection.UserEmail != normalizeEmail(userEmail) {
		return ErrNotFound
	}
	return s.mcpConnections.Delete([]byte(id), nil)
}
