package dpass

import "encoding/json"

// JSON is an unncrypted JSON representation of the generation options
func (g *GenOpts) JSON() ([]byte, error) {
	return json.Marshal(g)
}

func FromJSON(d []byte) (*GenOpts, error) {
	g := &GenOpts{}
	return g, json.Unmarshal(d, g)
}

// BlobIndex returns the index string which can identify an encrypted
// options blob. The first 22 characters are the base64 sha512_128 sum of the
// domain name. The remaining 6 characters are a hash of all the options to
// uniquely identify this entry for the domain.
func (g *GenOpts) BlobIndex() string {
	return ""
}
