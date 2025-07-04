/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transaction

import (
	"bytes"
	"encoding/json"

	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/errors"
	"github.com/hyperledger-labs/fabric-smart-client/pkg/utils/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
)

type Envelope struct {
	protoEnvelope *common.Envelope
	txID          string
	nonce         []byte
	creator       []byte
	results       []byte
}

func NewEmptyEnvelope() *Envelope {
	return &Envelope{
		protoEnvelope: &common.Envelope{},
	}
}

func NewEnvelope(txID string, nonce, creator, results []byte, protoEnvelope *common.Envelope) *Envelope {
	return &Envelope{
		protoEnvelope: protoEnvelope,
		txID:          txID,
		nonce:         bytes.Clone(nonce),
		creator:       bytes.Clone(creator),
		results:       bytes.Clone(results),
	}
}

func (e *Envelope) TxID() string {
	return e.txID
}

func (e *Envelope) Nonce() []byte {
	return e.nonce
}

func (e *Envelope) Creator() []byte {
	return e.creator
}

func (e *Envelope) Results() []byte {
	return e.results
}

func (e *Envelope) Bytes() ([]byte, error) {
	return protoutil.Marshal(e.protoEnvelope)
}

func (e *Envelope) FromBytes(raw []byte) error {
	var err error
	e.protoEnvelope, err = protoutil.UnmarshalEnvelope(raw)
	if err != nil {
		return err
	}

	return nil
}

func (e *Envelope) String() string {
	s, err := json.MarshalIndent(e.protoEnvelope, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(s)
}

func (e *Envelope) Envelope() *common.Envelope {
	return e.protoEnvelope
}

type UnpackedEnvelope struct {
	TxID     string
	Results  []byte
	Envelope []byte
}

func UnpackEnvelopeFromBytes(raw []byte) (*UnpackedEnvelope, int32, error) {
	env := &common.Envelope{}
	if err := proto.Unmarshal(raw, env); err != nil {
		return nil, -1, err
	}
	return UnpackEnvelope(env)
}

func GetChannelHeaderType(raw []byte) (common.HeaderType, error) {
	env := &common.Envelope{}
	if err := proto.Unmarshal(raw, env); err != nil {
		return -1, err
	}
	payl, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		return -1, errors.Wrap(err, "failed to unmarshal payload")
	}

	chdr, err := protoutil.UnmarshalChannelHeader(payl.Header.ChannelHeader)
	if err != nil {
		return -1, errors.Wrap(err, "failed to unmarshal channel header")
	}

	return common.HeaderType(chdr.Type), nil
}

func UnpackEnvelope(env *common.Envelope) (*UnpackedEnvelope, int32, error) {
	return UnpackEnvelopePayload(env.Payload)
}

func UnpackEnvelopePayload(payloadRaw []byte) (*UnpackedEnvelope, int32, error) {
	payl, err := protoutil.UnmarshalPayload(payloadRaw)
	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to unmarshal payload")
	}

	chdr, err := protoutil.UnmarshalChannelHeader(payl.Header.ChannelHeader)
	if err != nil {
		return nil, -1, errors.Wrap(err, "failed to unmarshal channel header")
	}

	// validate the payload type
	if common.HeaderType(chdr.Type) != common.HeaderType_MESSAGE {
		return nil, chdr.Type, errors.Errorf("only HeaderType_MESSAGE Transactions are supported, provided type %d", chdr.Type)
	}

	return &UnpackedEnvelope{
		TxID:    chdr.TxId,
		Results: payl.Data,
	}, chdr.Type, nil
}

func (u *UnpackedEnvelope) ID() string {
	return u.TxID
}
