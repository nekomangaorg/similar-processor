# The Catalyst's Journal

## 2025-02-12 - Memory Leak Prevention via Iterator Streaming and O(N^2) Query Scaling
**Learning:** Returning massive rows from the database as `[]internal.DbManga` consumes huge amounts of memory and causes unnecessary GC pressure. Furthermore, querying all DB items using sequential `LIMIT 100 OFFSET ?` causes the DB engine to linearly scan and discard rows for every query, scaling `O(N^2)` on execution time.
**Action:** Expose bulk record fetching via Go 1.22 `iter.Seq` which yields rows to an `O(1)` memory iterator. Replaced quadratic sequence chunks with a single sequential `SELECT` statement mapping results into size-100 batch arrays in pure Go.

## 2025-05-23 - Prepared Statement Lifecycle and Test Isolation
**Learning:** Using package-level prepared statements with `sync.Once` significantly reduces SQL parsing overhead in hot paths, but creates a hard dependency on the global database connection.
**Action:** Always provide a reset function (e.g., `resetExistsInDatabaseStmt`) that closes the statement and resets the `sync.Once` flag, and ensure it is called in test cleanup (`tb.Cleanup`) to prevent cross-test contamination when the database is swapped.
