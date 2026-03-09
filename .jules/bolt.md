## 2024-05-24 - [Optimizing Sparse Cosine Similarity]
**Learning:** `pairwise.CosineSimilarity` (and generic interface access like `mat.Vector.At()`) introduces significant overhead when computing similarity for large sparse datasets. By accessing the underlying `RawVector` data (indices and values) of `*sparse.Vector` and implementing a custom dot product that assumes sorted indices (O(NNZ)), we achieved a >2.5x speedup per operation. Additionally, pre-calculating L2 norms avoids redundant O(N) calculations in the inner loop.
**Action:** When working with `gonum` or `sparse` libraries in tight loops, check if you can type-assert to concrete types to bypass interface method dispatch overhead and access raw data structures for manual optimization.

## 2024-05-18 - [Word Counting Without Allocations]
**Learning:** `len(strings.Split(s, " "))` generates massive unnecessary heap allocations for word counting. `strings.Count(s, " ") + 1` is faster but can be inaccurate with multiple or leading/trailing spaces.
**Action:** For true word counting without allocations, implement a stateful counter iterating through bytes to correctly handle spaces, or use `strings.Count` when strictly replacing `strings.Split` behavior where identical (even if flawed) logic is required.
