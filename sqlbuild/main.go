package sqlbuild

/**
 * Utility methods for creating SQL query's
 */
import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Lazy create key and values for SQL Insert
func Keys(keys []string) (string, string) {
	var outKey string
	var outVal string
	for _, key := range keys {
		outKey = outKey + "`,`" + key
		outVal = outVal + ",?"
	}
	outKey = outKey + "`"
	return outKey[2:], outVal[1:]
}

// Lazy create key=value for SQL Update
func Keys2(keys []string) string {
	var out string
	for _, key := range keys {
		out = out + ",`" + key + "` = ?"
	}
	return out[1:]
}

// Convert ""-string into NULL
func AllowNullString(s string) *string {
	if len(s) == 0 {
		return nil
	}
	return &s
}

// Convert "0" into NULL
func AllowNullZero(s string) *string {
	if s == "0" {
		return nil
	}
	return &s
}

// Convert 0 into NULL
func AllowNullInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// Create LIMIT for paginating with MySQL
func Paginate(page int64, size int64) string {
	if page == 0 {
		return strconv.FormatInt(size, 10)
	}
	return strconv.FormatInt(page*size, 10) + "," + strconv.FormatInt(size, 10)
}

func SaveQuery(table string, keys []string, id int64) string {
	pos := strings.LastIndex(table, "_")
	prefix := table
	if pos != -1 {
		prefix = table[pos+1:]
	}

	if id == 0 {
		a := prefix + "_date_added"
		k, v := Keys(keys)
		return fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES(%s, %d)", table, k, a, v, time.Now().Unix())
	}

	a := prefix + "_date_updated"
	kv := Keys2(keys)
	priKey := table + "_id"
	return fmt.Sprintf("UPDATE %s SET %s, %s=%d WHERE %s = %d", table, kv, a, time.Now().Unix(), priKey, id)
}
