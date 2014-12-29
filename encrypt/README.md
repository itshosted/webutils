Simplify AES encrypt/decrypt.

```
func EncryptBase64(enc string, iv string, in interface{}) (string, error)
func DecryptBase64(enc string, iv string, in string, out interface{}) error
```

This code is used in the session-package to support client-side cookies. Have a
peek there for an example implementation. https://github.com/xsnews/webutils/tree/master/session

```
import (
  "github.com/xsnews/encrypt"
)

...

str, err := encrypt.EncryptBase64("aes", "32charstring____________________", instance)
```
