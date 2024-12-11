package trie

import (
	"bytes"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
)

// VerifyProof verifies a proof against a root hash and a key.
//
// If the proof if valid it returns a nil error and
// - value, the value associated with the key, it is non-nil if the key exists in the trie (inclusion proof) and nil otherwise (exclusion proof)
// - longestPrefix, the longest prefix of the key that exists in the trie.
// - lastNode, the node (collapsed) at longestPrefix that either proves the inclusion or exclusion of the key.
func VerifyProof(root gethcommon.Hash, key []byte, proofDB ethdb.KeyValueReader) (value []byte, err error) {
	value, _, _, err = VerifyProofWithReporting(root, key, proofDB, nil, nil)
	return value, err
}

func VerifyProofWithLastNode(root gethcommon.Hash, key []byte, proofDB ethdb.KeyValueReader) (value, longestPrefix, lastNode []byte, err error) {
	return VerifyProofWithReporting(root, key, proofDB, nil, nil)
}

func VerifyProofWithReporting(
	root gethcommon.Hash,
	key []byte,
	proofDB ethdb.KeyValueReader,
	reportNode func(hash gethcommon.Hash, key, node []byte),
	reportLeaf func(parentHash gethcommon.Hash, leaf []byte),
) (value, longestPrefix, lastNode []byte, err error) {
	key = keybytesToHex(key)
	wantHash := root
	nodeKey := []byte{}

	hasher := newHasher(false)
	defer returnHasherToPool(hasher)

	for i := 0; ; i++ {
		buf, _ := proofDB.Get(wantHash[:])
		if buf == nil {
			return nil, nil, nil, fmt.Errorf("proof node %d (hash %064x) missing", i, wantHash)
		}

		n, err := decodeNode(wantHash[:], buf)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("bad proof node %d: %v", i, err)
		}

		if reportNode != nil {
			reportNode(wantHash, nodeKey, buf)
		}

		keyLeft, keyVisited, cld := getCustom(n, key, true)
		nodeKey = append(nodeKey, keyVisited...)
		switch cld := cld.(type) {
		case nil:
			// The trie doesn't contain the key.
			collapsed, _ := hasher.proofHash(n)
			return nil, longestPrefix, nodeToBytes(collapsed), nil
		case hashNode:
			key = keyLeft
			longestPrefix = append(longestPrefix, keyVisited...)
			copy(wantHash[:], cld)
		case valueNode:
			collapsed, _ := hasher.proofHash(n)
			if reportLeaf != nil {
				reportLeaf(wantHash, cld)
			}
			return cld, longestPrefix, nodeToBytes(collapsed), nil
		}
	}
}

// getCustom returns the child of the given node. Return nil if the
// node with specified key doesn't exist at all.
//
// There is an additional flag `skipResolved`. If it's set then
// all resolved nodes won't be returned.
func getCustom(tn node, key []byte, skipResolved bool) (keyLeft, keyVisited []byte, n node) {
	keyLeft = key
	keyVisited = []byte{}
	for {
		switch n := tn.(type) {
		case *shortNode:
			if !bytes.HasPrefix(keyLeft, n.Key) {
				return keyLeft, keyVisited, nil
			}
			tn = n.Val
			keyVisited = n.Key[:]
			keyLeft = keyLeft[len(n.Key):]
			if !skipResolved {
				return keyLeft, keyVisited, tn
			}
		case *fullNode:
			tn = n.Children[keyLeft[0]]
			keyVisited = []byte{keyLeft[0]}
			keyLeft = keyLeft[1:]
			if !skipResolved {
				return keyLeft, keyVisited, tn
			}
		case hashNode:
			return keyLeft, keyVisited, n
		case nil:
			return keyLeft, keyVisited, nil
		case valueNode:
			return nil, keyVisited, n
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
	}
}
