package store

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/syndtr/goleveldb/leveldb"
)

func (s *Store) CreateAPIToken(token model.APIToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	exists, err := s.apiTokens.Has([]byte(token.ID), nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrExists
	}
	return putJSON(s.apiTokens, []byte(token.ID), token)
}

func (s *Store) GetAPIToken(id string) (model.APIToken, error) {
	var token model.APIToken
	if err := getJSON(s.apiTokens, []byte(id), &token); err != nil {
		return model.APIToken{}, err
	}
	return token, nil
}

func (s *Store) ListAPITokens(userEmail string) ([]model.APIToken, error) {
	iterator := s.apiTokens.NewIterator(nil, nil)
	defer iterator.Release()
	tokens := make([]model.APIToken, 0)
	for iterator.Next() {
		var token model.APIToken
		if err := json.Unmarshal(iterator.Value(), &token); err != nil {
			return nil, err
		}
		if strings.EqualFold(token.UserEmail, userEmail) {
			tokens = append(tokens, token)
		}
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(tokens, func(i, j int) bool { return tokens[i].CreatedAt.Before(tokens[j].CreatedAt) })
	return tokens, nil
}

func (s *Store) DeleteAPIToken(id, userEmail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, err := s.GetAPIToken(id)
	if err != nil {
		return err
	}
	if !strings.EqualFold(token.UserEmail, userEmail) {
		return ErrNotFound
	}
	return s.apiTokens.Delete([]byte(id), nil)
}

func (s *Store) CreateWebhookSubscription(subscription model.WebhookSubscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	iterator := s.webhookSubscriptions.NewIterator(nil, nil)
	defer iterator.Release()
	for iterator.Next() {
		var existing model.WebhookSubscription
		if err := json.Unmarshal(iterator.Value(), &existing); err != nil {
			return err
		}
		if existing.CallbackURL == subscription.CallbackURL {
			return ErrExists
		}
	}
	if err := iterator.Error(); err != nil {
		return err
	}
	return putJSON(s.webhookSubscriptions, []byte(subscription.ID), subscription)
}

func (s *Store) GetWebhookSubscription(id string) (model.WebhookSubscription, error) {
	var subscription model.WebhookSubscription
	if err := getJSON(s.webhookSubscriptions, []byte(id), &subscription); err != nil {
		return model.WebhookSubscription{}, err
	}
	return subscription, nil
}

func (s *Store) ListWebhookSubscriptions() ([]model.WebhookSubscription, error) {
	iterator := s.webhookSubscriptions.NewIterator(nil, nil)
	defer iterator.Release()
	subscriptions := make([]model.WebhookSubscription, 0)
	for iterator.Next() {
		var subscription model.WebhookSubscription
		if err := json.Unmarshal(iterator.Value(), &subscription); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *Store) CreateWebhookDeliveries(deliveries []model.WebhookDelivery) error {
	batch := new(leveldb.Batch)
	for _, delivery := range deliveries {
		encoded, err := json.Marshal(delivery)
		if err != nil {
			return err
		}
		batch.Put([]byte(delivery.ID), encoded)
	}
	return s.webhookDeliveries.Write(batch, nil)
}

func (s *Store) ListDueWebhookDeliveries(now time.Time) ([]model.WebhookDelivery, error) {
	iterator := s.webhookDeliveries.NewIterator(nil, nil)
	defer iterator.Release()
	deliveries := make([]model.WebhookDelivery, 0)
	for iterator.Next() {
		var delivery model.WebhookDelivery
		if err := json.Unmarshal(iterator.Value(), &delivery); err != nil {
			return nil, err
		}
		if !delivery.NextAttemptAt.After(now) {
			deliveries = append(deliveries, delivery)
		}
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(deliveries, func(i, j int) bool { return deliveries[i].NextAttemptAt.Before(deliveries[j].NextAttemptAt) })
	return deliveries, nil
}

func (s *Store) ListWebhookDeliveries() ([]model.WebhookDelivery, error) {
	return s.ListDueWebhookDeliveries(time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC))
}

func (s *Store) PutWebhookDelivery(delivery model.WebhookDelivery) error {
	return putJSON(s.webhookDeliveries, []byte(delivery.ID), delivery)
}

func (s *Store) DeleteWebhookDelivery(id string) error {
	return s.webhookDeliveries.Delete([]byte(id), nil)
}
