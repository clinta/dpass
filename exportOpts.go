package dpass

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"
)

// JSON is an unncrypted JSON representation of the generation options
func (g *GenOpts) JSON() ([]byte, error) {
	return json.Marshal(g)
}

func FromJSON(d []byte) (*GenOpts, error) {
	g := &GenOpts{}
	return g, json.Unmarshal(d, g)
}

// Returns the sha512_256 of the scrypt hash and the domain name, used both for
// generating the blobIndex and the blob encryption key
func (g *GenOpts) blobKey() ([32]byte, error) {
	if g.mpHash == [64]byte{} {
		return [32]byte{}, fmt.Errorf("No password has been hashed yet.")
	}
	seedSrc := append(g.mpHash[:], []byte(g.Domain)...)
	dh := sha512.Sum512_256(seedSrc)
	return dh, nil
}

func (g *GenOpts) jsonKey() (string, error) {
	j, err := g.JSON()
	if err != nil {
		return "", err
	}
	jh := sha512.Sum512_256(j)
	return base32.HexEncoding.EncodeToString(jh[:5]), nil
}

func (g *GenOpts) blobIndexPrefix() (string, error) {
	dh, err := g.blobKey()
	if err != nil {
		return "", err
	}
	dh = sha512.Sum512_256(dh[:])
	// base32 extended works on case insensitive filesystems
	return base32.HexEncoding.EncodeToString(dh[:10]), nil
}

// Given a domain and password, return the blob index prefix which
// can be used by an interface to look up all blobs for that domain
func BlobIndexPrefix(dom string, pw []byte) (string, error) {
	g := NewGenOpts("", dom)
	if err := g.HashPw(pw); err != nil {
		return "", err
	}
	return g.blobIndexPrefix()
}

// BlobIndex returns the index string which can identify an encrypted
// options blob. The first 22 characters are the base32 double sha512_128 sum of the
// domain name and mphash.
// The remaining 6 characters are a hash of the json encoded options to
// uniquely identify this entry for the domain.
// This makes searching for entries for a given domain possible before decryption.
func (g *GenOpts) BlobIndex() (string, error) {
	bp, err := g.blobIndexPrefix()
	if err != nil {
		return "", err
	}
	jk, err := g.jsonKey()
	if err != nil {
		return "", err
	}
	s := bp + jk
	return s, nil
}

// Blob returns a base32 encoded encrypted blob of the json encoded options.
// First the json is compressed, then encrypted using the sha512_256 sum of the
// scrypt master password hash + the domain name. This makes the encryption key
// unique for each domain, but allows decrypting all entries for a single domain
// with a single key so that an interface can allow choosing from entries in a
// domain.
func (g *GenOpts) Blob() (string, error) {
	// Get the json
	j, err := g.JSON()
	if err != nil {
		return "", err
	}

	// Compress the json
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(j)
	w.Close()
	jz := b.Bytes()

	// Random nonce
	var n [24]byte
	rand.Read(n[:])

	// Get the key
	dh, err := g.blobKey()
	if err != nil {
		return "", err
	}

	// Seal it with the nonce prepended
	out := make([]byte, len(n))
	copy(out, n[:])
	out = secretbox.Seal(out, jz, &n, &dh)

	// base64 encode it
	s := base64.URLEncoding.EncodeToString(out)
	return s, nil
}
