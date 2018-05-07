gh-asset
========

Download release assets from GitHub, with configurable retries.

Handy for fetching tarballs from Travis or other CI systems that might have
flaky network connections or that hit GitHub's API request rate limit.

Example:

```
gh-asset -d /tmp/ -x CanonicalLtd sqlite "sqlite-amd64--enable-debug-.*.tar.gz"
```
