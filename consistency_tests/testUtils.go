package ctests

import "fmt"

func StrLit(s string) string {
	return fmt.Sprintf("`%s`", s)
}

type TestParam struct {
	MasterPass string
	JSONOpts   string
}
