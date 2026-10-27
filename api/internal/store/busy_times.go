package store

import (
	"encoding/json"
	"sort"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/syndtr/goleveldb/leveldb"
)

func (s *Store) CreateBusyTime(busyTime model.BusyTime) error {
	return putJSON(s.busyTimes, []byte(busyTime.ID), busyTime)
}

func (s *Store) GetBusyTime(id string) (model.BusyTime, error) {
	var busyTime model.BusyTime
	if err := getJSON(s.busyTimes, []byte(id), &busyTime); err != nil {
		return model.BusyTime{}, err
	}
	return busyTime, nil
}

func (s *Store) ListBusyTimes() ([]model.BusyTime, error) {
	iterator := s.busyTimes.NewIterator(nil, nil)
	defer iterator.Release()
	busyTimes := make([]model.BusyTime, 0)
	for iterator.Next() {
		var busyTime model.BusyTime
		if err := json.Unmarshal(iterator.Value(), &busyTime); err != nil {
			return nil, err
		}
		busyTimes = append(busyTimes, busyTime)
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(busyTimes, func(i, j int) bool { return busyTimes[i].Start.Before(busyTimes[j].Start) })
	return busyTimes, nil
}

func (s *Store) DeleteBusyTime(id string) error {
	return s.busyTimes.Delete([]byte(id), nil)
}

func (s *Store) DeleteBusyTimesForUser(email string) error {
	busyTimes, err := s.ListBusyTimes()
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	for _, busyTime := range busyTimes {
		if busyTime.UserEmail == normalizeEmail(email) {
			batch.Delete([]byte(busyTime.ID))
		}
	}
	return s.busyTimes.Write(batch, nil)
}
