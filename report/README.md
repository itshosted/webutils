Simple logging abstraction.
Note: This lib is high on the refactor-list as syslog is a better system.

Write to path/(activity|error).log
activity.log contains regular info supplied by Msg and Debug(isVerbose)
error.log contains Err supplied info.


If isVerbose is set to true all logging text is appended to stdout and stderr
besides the log-file for easily debugging.


```
func Init(prefix string, path string, isVerbose bool) error
func Close()

func Debug(s string)
func Msg(s string)
func Err(e error)
```
