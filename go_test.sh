#go test -v


rm -v tests/sqlite.db || true
go test -v "vd-alpha/packages/collection"

