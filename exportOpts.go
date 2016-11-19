package dpass

import (
	"encoding/json"
	"io"
)

// JSON is an unncrypted JSON representation of the generation options
func (g *GenOpts) JSON() ([]byte, error) {
	return json.Marshal(g)
}

func FromJSON(d []byte) (*GenOpts, error) {
	g := &GenOpts{}
	return g, json.Unmarshal(d, g)
}

func DecodeJSON(r io.Reader) (*GenOpts, error) {
	g := &GenOpts{}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	return g, decoder.Decode(g)
}
