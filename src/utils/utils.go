package utils

import (
	"fmt"
	"strings"
	"time"
)

type ProposalKV struct {
	Key   string `datastore:"key"`
	Value int    `datastore:"value"`
}

type InnerKV struct {
	Key   string `datastore:"key"`
	Value bool   `datastore:"value"`
}

type OuterKV struct {
	Key   string    `datastore:"key"`
	Value []InnerKV `datastore:"value"`
}

func FormatTimeLeft(endTime time.Time) string {
	currentTime := time.Now()
	duration := endTime.Sub(currentTime)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	return fmt.Sprintf("%d days %02d hours", days, hours)
}

func GenerateProposalDetailURL(baseURL, chainName, proposalID string) string {
	return fmt.Sprintf("%s/%s/proposals/%s", baseURL, strings.ToLower(chainName), proposalID)
}

func MapToSlice(m map[string]int) []ProposalKV {
	var kvs []ProposalKV
	for k, v := range m {
		kvs = append(kvs, ProposalKV{Key: k, Value: v})
	}
	return kvs
}

func SliceToMap(kvs []ProposalKV) map[string]int {
	m := make(map[string]int)
	for _, kv := range kvs {
		m[kv.Key] = kv.Value
	}
	return m
}

func MapToNestedSlice(m map[string]map[string]bool) []OuterKV {
	var kvs []OuterKV
	for k, v := range m {
		innerKvs := make([]InnerKV, 0, len(v))
		for innerK, innerV := range v {
			innerKvs = append(innerKvs, InnerKV{Key: innerK, Value: innerV})
		}
		kvs = append(kvs, OuterKV{Key: k, Value: innerKvs})
	}
	return kvs
}

func NestedSliceToMap(kvs []OuterKV) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for _, kv := range kvs {
		innerMap := make(map[string]bool)
		for _, innerKv := range kv.Value {
			innerMap[innerKv.Key] = innerKv.Value
		}
		m[kv.Key] = innerMap
	}
	return m
}
