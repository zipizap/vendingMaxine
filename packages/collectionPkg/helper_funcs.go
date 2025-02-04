package collectionPkg

import (
	"fmt"
	"strings"
)

var colIdPrefix = "ColID-"

// colID_2_dbColID converts a collection ID (ex: "ColID-1234") to a database collection ID (ex: 1234 uint)
func colID_2_dbColID(colID string) (dbColID uint, err error) {
	// colID is of the form "colId-1234"
	if !strings.HasPrefix(colID, colIdPrefix) {
		return 0, fmt.Errorf("invalid collection ID prefix")
	}
	_, err = fmt.Sscanf(colID, colIdPrefix+"%d", &dbColID)
	if err != nil {
		return 0, err
	}
	return dbColID, nil
}

func dbColID_2_colID(dbColID uint) string {
	return fmt.Sprintf("%s%d", colIdPrefix, dbColID)
}
