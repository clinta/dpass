package main

const paramBegin = `package ctests

var TestParams = []TestParam{
`

const paramEntry = `	{
		MasterPass:  %s,
		JSONOpts: %s,
	},
`

const paramEnd = `}
`
