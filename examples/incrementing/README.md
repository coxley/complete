# Summary

While this is a contrived example, it demontrates using context from other parts of the
command-line to influence tab suggestions.

Each command in this CLI takes a `--num` flag. When using tab complete, the suggested
value will be a `+1` increment of the previous one.

```
# Will suggest 2
root --num 1 child --num <TAB>

# Will suggest 3
root --num 1 child --num 2 sub-child --num <TAB>
```

Read `main_test.go` to see all of the outcomes.

# Testing

To test the completion live, build the binary and 


```
go build -o /tmp/increment
COMP_INSTALL=1 /tmp/increment

/tmp/increment --num <TAB>
COMP_UNINSTALL=1 /tmp/increment
```
