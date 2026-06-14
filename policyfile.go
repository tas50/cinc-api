package cinc

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParsePolicyfileLock unmarshals the bytes of a Policyfile.lock.json into a
// PolicyRevision, which models the lock's structure (name, run lists, cookbook
// locks, attributes, solution dependencies). Use it to inspect a lock — for
// example to discover which cookbooks must be fetched before a push.
//
// To deploy a lock, prefer Policies.PushRevision with the original bytes: a
// lock can carry fields PolicyRevision does not model (such as a cookbook
// lock's "origin"), so re-marshalling a parsed value would drop them.
func ParsePolicyfileLock(data []byte) (*PolicyRevision, error) {
	var lock PolicyRevision
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("cinc: parse policyfile lock: %w", err)
	}
	if lock.Name == "" {
		return nil, fmt.Errorf("cinc: policyfile lock is missing a policy name")
	}
	return &lock, nil
}

// LoadPolicyfileLock reads and parses a Policyfile.lock.json from disk. It
// returns the parsed lock alongside the original bytes, which callers pass
// verbatim to Policies.PushRevision so no lock fields are lost on the round
// trip.
func LoadPolicyfileLock(path string) (*PolicyRevision, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("cinc: read policyfile lock: %w", err)
	}
	lock, err := ParsePolicyfileLock(data)
	if err != nil {
		return nil, nil, err
	}
	return lock, data, nil
}
