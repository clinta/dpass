To generate a consistency test for new version of a generator run:

```bash
go run goTest.go template.go | gofmt > v2tests.go
```
