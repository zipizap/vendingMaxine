#go test -v


rm -v tests/sqlite.db || true
go test -v "vendingMaxine/packages/collection"

