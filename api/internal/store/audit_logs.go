package store

import (
	"encoding/json"
	"fmt"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/syndtr/goleveldb/leveldb"
)

const AuditLogsTableName = "Auditlogs"

func (s *Store) AppendAuditLog(entry model.AuditLog, maxItems int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := []byte(fmt.Sprintf("%020d/%s", entry.CreatedAt.Unix(), entry.ID))
	exists, err := s.auditLogs.Has(key, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrExists
	}
	encoded, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	iterator := s.auditLogs.NewIterator(nil, nil)
	defer iterator.Release()
	batch := new(leveldb.Batch)
	keys := make([][]byte, 0)
	for iterator.Next() {
		keys = append(keys, append([]byte(nil), iterator.Key()...))
	}
	if err := iterator.Error(); err != nil {
		return err
	}
	for _, existingKey := range keys[:max(0, len(keys)-maxItems+1)] {
		batch.Delete(existingKey)
	}
	batch.Put(key, encoded)
	return s.auditLogs.Write(batch, nil)
}

func (s *Store) ListAuditLogs() ([]model.AuditLog, error) {
	iterator := s.auditLogs.NewIterator(nil, nil)
	defer iterator.Release()
	entries := make([]model.AuditLog, 0)
	for ok := iterator.Last(); ok; ok = iterator.Prev() {
		var entry model.AuditLog
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	return entries, nil
}
