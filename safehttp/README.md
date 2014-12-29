Require admin-privilege unless excluded.

Combining the middleware + session packages to add simple
user management by blocking everything unless a rule is defined.

```
safehttp.Add("/allowAnyOne", safehttp.Rule{false})
safehttp.Add("/allowLoggedIn", safehttp.Rule{true})
middleware.Add(safehttp.Use(true, "IVOf32chars____________________"))
http.Handle("/", middleware.Use(mux))
```
