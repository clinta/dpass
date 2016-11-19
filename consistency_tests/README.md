To generate a consistency test for new version of a generator run:

DO NOT ever regenerate tests for a previous version after that version has been
tagged and released. If those tests break then the generator must be incremented
to preserve backwards compatibility.

```bash
go run genTests/*.go > v1_test.go.tmp && mv v1_test.go.tmp v1_test.go
```

Testing parameters were generated using:

```bash
go run genParams/*.go > params.go.tmp && mv params.go.tmp params.go
```
