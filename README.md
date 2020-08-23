# goreq

Checks go dependencies against a predefined list of module requirements

1. Reads go.mod
2. Reads go.req (yaml formatted)
3. Returns error if any dependencies don't meet required version
4. Returns error if any dependencies are banned
4. Returns success if all dependencies are allowed
