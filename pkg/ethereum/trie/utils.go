package trie

import (
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
)

// StorageTrieKey returns the key used to store a slot in storage trie.
func StorageTrieKey(slot []byte) []byte {
	return crypto.Keccak256(slot)
}

// AccountTrieKey returns the key used to store an account in the account trie.
func AccountTrieKey(addr gethcommon.Address) []byte {
	return crypto.Keccak256(addr.Bytes())
}

// StorageTrieOwner returns the owner of the storage trie for a given account.
func StorageTrieOwner(addr gethcommon.Address) gethcommon.Hash {
	return crypto.Keccak256Hash(addr.Bytes())
}

// AccountTrieOwner returns the owner of the account trie.
func AccountTrieOwner() gethcommon.Hash {
	return gethcommon.Hash{}
}

// ProofsToDB stores a list of proofs in a database.
func FromBytesProof(proof [][]byte, db ethdb.KeyValueWriter) error {
	for _, node := range proof {
		hash := crypto.Keccak256(node)
		if err := db.Put(hash, node); err != nil {
			return fmt.Errorf("failed to store proof node %q: %w", node, err)
		}
	}
	return nil
}

// ProofsToDB stores a list of proofs in a database.
func FromHexProof(proof []string, db ethdb.KeyValueWriter) error {
	byteProofs := make([][]byte, len(proof))
	for i, node := range proof {
		byteProof, err := hexutil.Decode(node)
		if err != nil {
			return fmt.Errorf("failed to decode proof node %q: %w", node, err)
		}
		byteProofs[i] = byteProof
	}
	return FromBytesProof(byteProofs, db)
}
