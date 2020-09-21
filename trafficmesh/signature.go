package trafficmesh

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dgrijalva/jwt-go"
)

const signatureHeader = "X-NF-Mesh-Signature"

type SignatureDecoder struct {
	secret string
}

type SignaturePayload struct {
	jwt.StandardClaims
	SiteID    string `json:"sid,omitempty"`
	DeployID  string `json:"did,omitempty"`
	AccountID string `json:"aid,omitempty"`
	URL       string `json:"url,omitempty"`
	Remapped  bool   `json:"remapped,omitempty"`
}

func NewSignatureDecoder(secret string) *SignatureDecoder {
	return &SignatureDecoder{
		secret: secret,
	}
}

func (d *SignatureDecoder) DecodeSignature(req *http.Request) (*SignaturePayload, error) {
	if d.secret == "" || req.Header.Get(signatureHeader) == "" {
		return nil, nil
	}

	payload := new(SignaturePayload)
	token, err := jwt.ParseWithClaims(req.Header.Get(signatureHeader), payload, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || alg != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(d.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decode traffic mesh signature: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	payloadURL, err := url.Parse(payload.URL)
	if err != nil {
		return nil, err
	}
	if payloadURL.Host != req.Host {
		return nil, fmt.Errorf("token host %s doesn't match request host: %s", payloadURL.Host, req.Host)
	}
	if payloadURI, reqURI := payloadURL.RequestURI(), req.URL.RequestURI(); payloadURI != reqURI {
		return nil, fmt.Errorf("token uri %s doesn't match request uri: %s", payloadURI, reqURI)
	}

	return payload, nil
}
