/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vault

import (
	"bytes"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/core/generic/vault"
	"github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/types"
)

type CounterBasedVersionBuilder struct{}

func (*CounterBasedVersionBuilder) VersionedValues(rws *vault.ReadWriteSet, ns driver.Namespace, writes vault.NamespaceWrites, block driver.BlockNum, indexInBloc driver.TxNum) (map[driver.PKey]driver.VaultValue, error) {
	vals := make(map[driver.PKey]driver.VaultValue, len(writes))
	reads := rws.Reads[ns]

	for pkey, val := range writes {
		v, err := version(reads, pkey)
		if err != nil {
			return nil, err
		}
		vals[pkey] = driver.VaultValue{Raw: val, Version: v}
	}
	return vals, nil
}

func version(reads vault.NamespaceReads, pkey driver.PKey) (vault.Version, error) {
	// Search the corresponding read.
	v, ok := reads[pkey]
	if !ok {
		// this is a blind write, we should check the vault.
		// Let's assume here that a blind write always starts from version 0
		return Marshal(0), nil
	}

	if len(v) == 0 {
		return Marshal(0), nil
	}

	// parse the version as an integer, then increment it
	counter, err := Unmarshal(v)
	if err != nil {
		return nil, errors.Wrapf(err, "failed unmarshalling version for %s:%v", pkey, v)
	}
	return Marshal(counter + 1), nil
}

func (c *CounterBasedVersionBuilder) VersionedMetaValues(rws *vault.ReadWriteSet, ns driver.Namespace, writes vault.KeyedMetaWrites, block driver.BlockNum, indexInBloc driver.TxNum) (map[driver.PKey]driver.VaultMetadataValue, error) {
	vals := make(map[driver.PKey]driver.VaultMetadataValue, len(writes))
	reads := rws.Reads[ns]

	for pkey, val := range writes {
		v, err := version(reads, pkey)
		if err != nil {
			return nil, err
		}

		vals[pkey] = driver.VaultMetadataValue{Metadata: val, Version: v}
	}
	return vals, nil
}

type CounterBasedVersionComparator struct{}

func (*CounterBasedVersionComparator) Equal(v1, v2 driver.RawVersion) bool {
	return bytes.Equal(v1, v2)
}

func Marshal(v uint32) []byte {
	return types.VersionNumber(v).Bytes()
}

func Unmarshal(raw []byte) (uint32, error) {
	return uint32(types.VersionNumberFromBytes(raw)), nil //nolint:gosec
}
