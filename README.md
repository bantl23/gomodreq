# gomodreq

Checks go dependencies against a predefined list of module requirements

1. Reads go.mod
2. Reads .gomodreq.yml
3. Support URI Schemes file, http, https, and ssh
4. Returns error if any dependencies don't meet required version
5. Returns error if any dependencies are banned
6. Returns success if all dependencies are allowed
