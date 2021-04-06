// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// This is nothing but an arbitrary number for safety. Due to the nature of
// the data in this header we can't know an exact max size.
const MaxAuxBlockHeaderPayload = 2048

// AuxBlockHeader defines information about an AuxPoW block and is used in the
// block (MsgBlock) and headers (MsgHeaders) messages.
type AuxBlockHeader struct {
	// Coinbase transaction that is in the parent block, linking the AuxPOW block to its parent block
	ParentCoinbase MsgTx

	// Hash of the parent_block header
	ParentBlockHash chainhash.Hash

	// The merkle branch linking the coinbase_txn to the parent block's merkle_root
	CoinbaseBranch MerkleBranch

	// The merkle branch linking this auxiliary blockchain to the others, when used in a
	//merged mining setup with multiple auxiliary chains
	BlockchainBranch MerkleBranch

	// Parent block header
	ParentBlock ParentBlock
}

// Essentially a copy of BlockHeader but due to cyclic references we can't use that
type ParentBlock struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32

	// Hash of the previous block header in the block chain.
	PrevBlock chainhash.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot chainhash.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32
}

type MerkleBranch struct {
	// Individual hash in the branch; repeated branch_length number of times
	LinkHashes []*chainhash.Hash

	// Bitmask of which side of the merkle hash function the branch_hash element should go on. Zero means it
	// goes on the right, One means on the left. It is equal to the index of the starting hash within
	// the widest level of the merkle tree for this merkle branch.
	BranchSidesBitmask int32
}

func (pb *ParentBlock) ToBlockHeader() *BlockHeader {
	return &BlockHeader{
		Version:    pb.Version,
		PrevBlock:  pb.PrevBlock,
		MerkleRoot: pb.MerkleRoot,
		Timestamp:  pb.Timestamp,
		Bits:       pb.Bits,
		Nonce:      pb.Nonce,
		AuxData:    AuxBlockHeader{},
	}
}

func (pb *ParentBlock) BtcDecode(r io.Reader) error {
	return readElements(r, &pb.Version, &pb.PrevBlock, &pb.MerkleRoot,
		(*uint32Time)(&pb.Timestamp), &pb.Bits, &pb.Nonce)
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding block headers stored to disk, such as in a
// database, as opposed to decoding block headers from the wire.
func (h *AuxBlockHeader) BtcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	return readAuxBlockHeader(r, pver, h)
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding block headers to be stored to disk, such as in a
// database, as opposed to encoding block headers for the wire.
func (h *AuxBlockHeader) BtcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	return writeAuxBlockHeader(w, pver, h)
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *AuxBlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	return readAuxBlockHeader(r, 0, h)
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *AuxBlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	return writeAuxBlockHeader(w, 0, h)
}

// NewBlockHeader returns a new BlockHeader using the provided version, previous
// block hash, merkle root hash, difficulty bits, and nonce used to generate the
// block with defaults for the remaining fields.
func NewAuxBlockHeader(version int32, prevHash, merkleRootHash *chainhash.Hash,
	bits uint32, nonce uint32) *AuxBlockHeader {

	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	return &AuxBlockHeader{
		// TODO finish this
		ParentBlock: ParentBlock{
			Version:    version,
			PrevBlock:  *prevHash,
			MerkleRoot: *merkleRootHash,
			Timestamp:  time.Unix(time.Now().Unix(), 0),
			Bits:       bits,
			Nonce:      nonce,
		},
	}
}

// readBlockHeader reads a bitcoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
func readAuxBlockHeader(r io.Reader, pver uint32, bh *AuxBlockHeader) error {
	bh.ParentCoinbase.BtcDecode(r, pver, BaseEncoding)

	_ = readElements(r, &bh.ParentBlockHash)

	count, err := ReadVarInt(r, pver)
	if err != nil {
		return err
	}
	coinbaseLinkHashes := make([]chainhash.Hash, count)
	bh.CoinbaseBranch.LinkHashes = make([]*chainhash.Hash, 0, count)
	for i := uint64(0); i < count; i++ {
		hash := &coinbaseLinkHashes[i]
		err := readElement(r, hash)
		if err != nil {
			return err
		}
		bh.CoinbaseBranch.LinkHashes = append(bh.CoinbaseBranch.LinkHashes, hash)
	}
	_ = readElement(r, bh.CoinbaseBranch.BranchSidesBitmask)

	count, err = ReadVarInt(r, pver)
	if err != nil {
		return err
	}
	blockchainLinkHashes := make([]chainhash.Hash, count)
	bh.BlockchainBranch.LinkHashes = make([]*chainhash.Hash, 0, count)
	for i := uint64(0); i < count; i++ {
		hash := &blockchainLinkHashes[i]
		err := readElement(r, hash)
		if err != nil {
			return err
		}
		bh.BlockchainBranch.LinkHashes = append(bh.BlockchainBranch.LinkHashes, hash)
	}
	_ = readElement(r, bh.BlockchainBranch.BranchSidesBitmask)

	bh.ParentBlock.BtcDecode(r)

	return nil
}

// writeBlockHeader writes a bitcoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
func writeAuxBlockHeader(w io.Writer, pver uint32, bh *AuxBlockHeader) error {
	bh.ParentCoinbase.BtcEncode(w, pver, BaseEncoding)

	_ = writeElements(w, &bh.ParentBlockHash)

	count := uint64(len(bh.CoinbaseBranch.LinkHashes))
	_ = WriteVarInt(w, pver, count)
	for i := uint64(0); i < count; i++ {
		_ = writeElements(w, pver, &bh.CoinbaseBranch.LinkHashes[i])
	}
	_ = writeElements(w, bh.CoinbaseBranch.BranchSidesBitmask)

	count = uint64(len(bh.BlockchainBranch.LinkHashes))
	_ = WriteVarInt(w, pver, count)
	for i := uint64(0); i < count; i++ {
		_ = writeElements(w, pver, &bh.BlockchainBranch.LinkHashes[i])
	}
	_ = writeElements(w, bh.BlockchainBranch.BranchSidesBitmask)

	_ = writeBlockHeader(w, pver, bh.ParentBlock.ToBlockHeader(), true)

	return nil
}
