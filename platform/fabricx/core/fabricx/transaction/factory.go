/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transaction

import (
	"context"

	driver2 "github.com/hyperledger-labs/fabric-smart-client/platform/common/driver"
	ftransaction "github.com/hyperledger-labs/fabric-smart-client/platform/fabric/core/generic/transaction"
	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
	"github.com/hyperledger/fabric-x-endorser/platform/fabricx/core/fabricx/committer/api/protoblocktx"
	"github.com/hyperledger/fabric/protoutil"
)

type TransactionFactory struct {
	fns     driver.FabricNetworkService
	adapter protoblocktx.Marshaller
}

func NewTransactionFactory(fns driver.FabricNetworkService, adapter protoblocktx.Marshaller) *TransactionFactory {
	return &TransactionFactory{fns: fns, adapter: adapter}
}

func (e *TransactionFactory) NewTransaction(ctx context.Context, channelName string, nonce, creator []byte, txID driver2.TxID, rawRequest []byte) (driver.Transaction, error) {
	ch, err := e.fns.Channel(channelName)
	if err != nil {
		return nil, err
	}

	if len(nonce) == 0 {
		nonce, err = ftransaction.GetRandomNonce()
		if err != nil {
			return nil, err
		}
	}
	if len(txID) == 0 {
		txID = protoutil.ComputeTxID(nonce, creator)
	}

	return &Transaction{
		ctx:        ctx,
		fns:        e.fns,
		adapter:    e.adapter,
		channel:    ch,
		TCreator:   creator,
		TNonce:     nonce,
		TTxID:      txID,
		TNetwork:   e.fns.Name(),
		TChannel:   channelName,
		TTransient: map[string][]byte{},
		// TODO: we need the correct values here
		// TODO: we need a mapper that link between classic strings to new sc namespace
		TChaincodeVersion: "1",
		TChaincode:        "iou",
	}, nil
}
