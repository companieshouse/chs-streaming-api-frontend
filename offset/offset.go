// Package offset provides functions for validating request offsets.
package offset

import (
	"errors"
	"strconv"

	"github.com/Shopify/sarama"

	"github.com/companieshouse/chs.go/kafka/client"
	"github.com/companieshouse/chs.go/log"
)

var (
	// ErrNegativeOffset is the error returned when the requested offset is less than zero.
	ErrNegativeOffset = errors.New("offset should be greater than zero")
	// ErrOutOfRange is the error returned when the requested offset is out of range.
	ErrOutOfRange = errors.New("requested offset is out of range")
)

// Offset - structure for offset
type Offset struct {
}

// Interface - interface for offset methods
type Interface interface {
	Parse(offset string) (int64, error)
	IsValid(brokerAddr []string, topic string, offset int64) error
}

// NewOffset - return a new Offset
func NewOffset() Interface {
	return &Offset{}
}

// Parse parses a string offset as an int64.
func (of *Offset) Parse(offset string) (int64, error) {
	if offset == "" {
		return 0, nil
	}

	o, err := strconv.Atoi(offset)
	if err != nil {
		return 0, err
	}
	if o < 0 {
		return 0, ErrNegativeOffset
	}

	return int64(o), nil
}

// IsValid validates that an offset is within a valid range for the provided topic and returns nil if a valid offset is passed.
func (of *Offset) IsValid(brokerAddr []string, topic string, offset int64) error {

	oldestOffset, err := OldestTopicOffset(brokerAddr, topic)
	if err != nil {
		return err
	}
	if offset < oldestOffset {
		log.Debug("Offset requested expired", log.Data{"offset_requested": offset, "oldest_offset": oldestOffset, "topic": topic})
		return ErrOutOfRange
	}

	newestOffset, err2 := client.TopicOffset(brokerAddr, topic)
	if err2 != nil {
		return err2
	}
	if offset > newestOffset {
		log.Debug("Offset requested not reached", log.Data{"offset_requested": offset, "newest_offset": newestOffset, "topic": topic})
		return ErrOutOfRange
	}

	return nil
}

// OldestTopicOffset Returns oldest offset
func OldestTopicOffset(brokerAddr []string, topic string) (int64, error) {
	client, err := sarama.NewClient(brokerAddr, sarama.NewConfig())
	if err != nil {
		return 0, err
	}
	offset, err := client.GetOffset(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return 0, err
	}
	return offset, nil
}
