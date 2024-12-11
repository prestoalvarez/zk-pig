# pkg/ethereum/trie

This package is largely forked from [go-ethereum v1.14.12](https://github.com/ethereum/go-ethereum/tree/v1.14.12/trie) 

It provides a modified version of Merkle Patricia Trie compatible with block execution on partial state.

It includes:

- A modified Merkle Patricia Trie (MPT) to handle specific behavior during node deletions.
- Utilities for preparing effective partial MPTs that support block execution.

This package is specifically designed for EVM block proving use cases.

## Warning

This MPT implementation is **not compatible** with Ethereum full nodes out of the box and would break most compatibility with the broader Ethereum ecosystem. Do not use this implementation without a thorough understanding of its approach and limitations.

---

## Architecture Note

### Background: MPT Node Reduction on Deletion

In a Merkle Patricia Trie, when a deletion leaves a branch node with only a single child, the branch node is replaced by a short node (either an extension or a leaf). This process is known as **branch node reduction**.

In the standard MPT implementation, during branch node reduction, the trie resolves the remaining child node. If the child node is a short node (i.e., an extension or a leaf), it is merged with the reduced branch node. Then if the parent of the branch node is also a short node, the reduced branch node is merged with the parent. This ensure the updated trie does not contain consecutive short nodes.

#### Branch node reduction in details

Reducing a branch node `brnch` at prefix key `brnch.prefix` consecutively to a deletion results in on of the following scenarios

- if the branch node has `N` children with `N > 2` => the branch node is replaced with a branch node with `N-1` children at the same prefix key `brnch.prefix`
- if the branch node has `N=2` children: the deleted child and the remaining child `cld_rem=brnch[nib_rem]` => the branch node is reduced 
    - if the remaining child `cld_rem` is not a short node => no merge is needed and `cld_rem` remains untouched
        - if the parent of the branch node `prnt=brnch.prnt` is not a short node => no merge is needed and `brnch` is replaced with a one-nibble short node `([nib_rem],cld_rem)` at same prefix key `brnch.prefix`
        - if the parent of the branch node is a is a short node `prnt=([nib_p_1,...,nib_p_p],brnch)` => a merge is needed, the parent short node is replaced with the extended short node `([nib_p_1,...,nib_p_p,nib_rem],cld_rem)` at same key prefix `prnt.prefix` and the branch node `brnch` at key prefix `brnch.prefix` is completly deleted and 
    - if the remaining child is a short node `cld_rem=([nib_c_1,...,nib_c_q],cld_rem.cld)` => a merge is needed
        - if the parent of the branch node is not a short node => `brnch` is replaced with the merged short node `([nib_rem,nib_c_1,...,nib_c_q],cld_rem.cld)` at the same prefix key `brnch.prefix` and the remaining child node `cld_rem` at prefix `cld_rm.prefix` is commpletly deleted
        - if the parent of the brnch node is a short node => the parent short node is replaced with the merged short node `([nib_p_1,...,nib_p_p,nib_rem,nib_c_1,...,nib_c_q],cld_rem.cld)` at the same prefix key `prnt.prefix` and both `brnch` at `brnch.prefix` and `cld_rem` at prefix `cld_rem.prefix` are completly deleted

##### Summary

| **Parent** | **Branch** | **Remaining Child** | **Replacement** | **Deletions** |
| --- | --- | --- | --- | --- |
| Any | `N>2` children | - | `brnch.prefix` -> Branch node with `N-1` children | None |
| Not Short Node | `N=2`children | Not short node | `brnch.prefix` -> Short node `([nib_rem],cld_rem)` | None |
| `([nib_p_1,...,nib_p_p],brnch)`| `N=2`children | Not short node | `prnt.prefix` -> Short node `([nib_p_1,...,nib_p_p,nib_rem],cld_rem.cld)` | `brnch` |
| Not Short Node | `N=2`children | `([nib_c_1,...,nib_c_q],cld_rem.cld)` | `brnch.prefix` ->  Short node `([nib_rem,nib_c_1,...,nib_c_q],cld_rem.cld)` | `cld_rem` |
| `([nib_p_1,...,nib_p_p],brnch)` | `N=2`children | `([nib_c_1,...,nib_c_q],cld_rem.cld)` | `prnt.prefix` -> Short node `([nib_p_1,...,nib_p_p, nib_rem,nib_c_1,...,nib_c_q],cld_rem.cld)` | `cld_rem` & `brnch` |

### Problem with Partial States

The standard MPT requires the remaining child node to be accessible during branch node reduction. While this works in a full-node environment, our use case involves **partial states**, which only include a subset of nodes. As such, we cannot always assume that the remaining child is accessible.

This creates a significant issue: the inability to compute the final state root after block execution from partial states in many scenarios.

### Modified Implementation

To address this limitation, we modified the MPT implementation. When reducing a branch node during a deletion, if the remaining child node is unavailable and cannot be resolved, the branch node is replaced with a **one-nibble short node** (either an extension or a leaf) representing the child. This approach assumes that if the remaining child is unavailable, it should be treated as though it is not a short node.

### Supplemental Technique: Pre-Injected Short Nodes

To further address this issue, we use a supplemental technique during pre-state preparation
1. **Hypothesize potential short-node reductions** based on post-state data.
2. **Pre-inject these hypothesized short nodes** into the pre-state.

This ensures that during reduction, any potential missing child short nodes is accounted for in advance.

#### Pre-State Preparation Workflow

To prepare a pre-state for deleting a key `k` during EVM execution:

1. Retrieve from an archival node:

- Pre-state proof: Inclusion proof `[n_1, n_2, ..., n_p]` for `k`.
- Post-state proof: Exclusion proof `[n'_1, n'_2, ..., n'_q]` for `k`.

2. Inject `[n_1, n_2, ..., n_p]` into the pre-state for basic access.
3. If `n'_q` is not a short node (i.e an extension or a leaf), no further action is needed.
4. If `n'_q` is a short node `([nib_1,nib_2,...,nib_l],cld)` compute all possible short nodes with shorten prefix key `([nib_2,...,nib_l],cld), ([nib_3,...,nib_l],cld)`, ...,`([nib_l],cld)`
5. Inject all short nodes in pre-state

This ensures that the pre-state can handle deletions, if the branch node reduction.

> **NOTE** This approach is actually a bit overkill, as `[n_1, n_2, ..., n_p]` and `[n'_1, n'_2, ..., n'_q]` would allow to infer the exact reduction scenario and missing child node of the reduced branch node thus injecting a single node in pre-state instead of a list. We leave it for future optimization.

### Resolution Scenarios

When the MPT resolves the remaining child during reduction, one of two scenarios occurs:

1. **Child Found:** The remaining child is one of the pre-injected short nodes. It is merged with the reduced branch node, following standard MPT behavior.
   
2. **Child Missing:** The remaining child is not one of the pre-injected short nodes which implies that the remaining is not a short node (or it would have been found). The modified MPT replaces the branch node with a one-nibble short node, which is the expected behavior for a non-short node child.

### Result

This modified implementation ensures compatibility with standard MPT behavior while enabling state root computation from partial states. In both scenarios, the resulting MPT structure aligns with that of the regular implementation.

### Considerations

This approach relies on access to **both pre-state and post-state data** to function effectively. It cannot be used to execute blocks in live mode. Instead, it assumes blocks have already been executed on full nodes, allowing the necessary pre- and post-state information to be retrieved.

---

## Implementation Note

This package is not a formal fork of `go-ethereum`. Instead, files were **copied and minimally modified** as follows:

- All files in this package are direct copies from `go-ethereum`, except those prefixed with `custom_*`.
- The `custom_*` files are included in this package to access internal methods that are not exported.
- The main MPT modification is in `trie.go` at [**lines 532-538**](./trie.go#L532).
- All original tests have been retained, except `TestMissingNodes()` in `trie_test.go`, which breaks due to the modification. This test has been commented out.
- Minor type adjustments were made to ensure the modified trie implements the same interface as the base `go-ethereum` trie, maintaining compatibility with other `go-ethereum` components (e.g., it can be injected into the EVM).
- We use a custom implementation of `VerifyProof` that passes all orginal go-ethereum tests