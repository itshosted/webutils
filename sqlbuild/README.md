Simplify INSERT/UPDATE and LIMIT.

SQL Query's are powerful and one shouldn't abstract them away in his/her code. The
only exception are the INSERT and UPDATE query's (which are simple and most of the
time not complex like SELECT).

Database design assumption
To simplify code the following assumptions are made:

- Table names are lower_case_separated
- The last separated text is used as prefix for columns (to keep JOIN-code simple)
  so table member_data has columns data_field1 and data_field2
- prefix_date_added and prefix_date_updated are columns that exist

This lib also provides Paginate() for simple pagination of results.

Abstracting INSERT/UPDATE
```
import (
  "github.com/xsnews/webutils/sqlbuild"
)

...

  // If 0 create INSERT else UPDATE
  id := 0

  q := sqlbuild.SaveQuery("table_name", []string{
    "name_field1", "name_field2"
  }, id)

  // txn is transaction from db.Begin()
  if _, e := txn.Exec(q,
    "value1", "value2"
  ); e != nil {
    return e
  }
  return nil
```

Pagination
```
limit := sqlbuild.Paginate(0, 5000)
// 5000
limit := sqlbuild.Paginate(1, 5000)
// 5001 - 10000
```
