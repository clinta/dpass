To generate a consistency test for new version of a generator run:

DO NOT ever regenerate tests for a previous version after that version has been
tagged and released. If those tests break then the generator must be incremented
to preserve backwards compatibility.

```bash
go run goTest.go template.go | gofmt > v2_test.go
```
