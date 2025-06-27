package model

import (
	"reflect"
	"testing"
	"time"
)

func TestMessageStructFields(t *testing.T) {
	now := time.Now()
	exp := now.Add(1 * time.Hour)
	msg := Message{
		ID:        "test-id",
		Encrypted: "encrypted-content",
		CreatedAt: now,
		ExpiresAt: exp,
	}

	if msg.ID != "test-id" {
		t.Errorf("expected ID to be 'test-id', got '%s'", msg.ID)
	}
	if msg.Encrypted != "encrypted-content" {
		t.Errorf("expected Encrypted to be 'encrypted-content', got '%s'", msg.Encrypted)
	}
	if !msg.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt to be '%v', got '%v'", now, msg.CreatedAt)
	}
	if !msg.ExpiresAt.Equal(exp) {
		t.Errorf("expected ExpiresAt to be '%v', got '%v'", exp, msg.ExpiresAt)
	}
}

func TestMessageBSONTags(t *testing.T) {
	msgType := Message{}
	bsonTags := map[string]string{
		"ID":        "_id",
		"Encrypted": "encrypted",
		"CreatedAt": "createdAt",
		"ExpiresAt": "expiresAt",
	}

	msgTypeType := reflect.TypeOf(msgType)
	for i := 0; i < msgTypeType.NumField(); i++ {
		field := msgTypeType.Field(i)
		expectedTag, ok := bsonTags[field.Name]
		if !ok {
			t.Errorf("unexpected field: %s", field.Name)
			continue
		}
		bsonTag := field.Tag.Get("bson")
		if bsonTag != expectedTag {
			t.Errorf("field %s: expected bson tag '%s', got '%s'", field.Name, expectedTag, bsonTag)
		}
	}
}
