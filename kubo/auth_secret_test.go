package kubo_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/NatoBoram/ipapm/kubo"
)

func TestDecodeAuthSecret_None(t *testing.T) {
	const token = "jYVZmR`euV.gV@jtY;pn=:Cf;wdH0Z7D})iSRStkHVNlrgd%i[>f,W5@0_nBy_G!i0P5JU)W"
	autoSecret, err := kubo.DecodeAuthSecret(token)
	if err != nil {
		t.Errorf("failed to decode bearer token: %s", err)
	}
	if autoSecret.Token != token {
		t.Errorf("Expected token to be %s, got %s", token, autoSecret.Token)
	}
}

func TestDecodeAuthSecret_Bearer(t *testing.T) {
	const token = "xT_Xw@:+1r-u(`NjF]Nh]<+v>L7v=-[,s5>)Sr2t%e!AuUqA^B]7ataVK^jHV>~f/<!pRKBx"
	bearer := fmt.Sprintf("bearer:%s", token)
	autoSecret, err := kubo.DecodeAuthSecret(bearer)
	if err != nil {
		t.Errorf("failed to decode bearer token: %s", err)
	}
	if autoSecret.Token != token {
		t.Errorf("Expected token to be %s, got %s", token, autoSecret.Token)
	}
}

func TestDecodeAuthSecret_Basic3(t *testing.T) {
	const username = "username"
	const password = "3|_sgD!ZIg0=;4uH-3n1vmr<!fUEQ6^zT.@I3^1Egxdp//?8Q~oFT:{a=OSLNH!D)vS4[=N["
	basicAuth := fmt.Sprintf("basic:%s:%s", username, password)

	autoSecret, err := kubo.DecodeAuthSecret(basicAuth)
	if err != nil {
		t.Errorf("failed to decode basic auth: %s", err)
	}
	if autoSecret.Username != username || autoSecret.Password != password {
		t.Errorf("Expected username to be %s and password to be %s, got username: %s, password: %s", username, password, autoSecret.Username, autoSecret.Password)
	}
}

func TestDecodeAuthSecret_Basic2(t *testing.T) {
	const username = "username"
	const password = "?8K:^R;?=]&2svQetZPQb~~tG9GTeX}OD.^mPB,;vVLG314sY?v6p(@uU|VJwzJ|O]>p|peu"
	combined := fmt.Appendf([]byte(""), "%s:%s", username, password)
	encoded := base64.StdEncoding.EncodeToString(combined)
	basicAuth := fmt.Sprintf("basic:%s", encoded)

	autoSecret, err := kubo.DecodeAuthSecret(basicAuth)
	if err != nil {
		t.Errorf("failed to decode basic auth: %s", err)
	}
	if autoSecret.Username != username || autoSecret.Password != password {
		t.Errorf("Expected username to be %s and password to be %s, got username: %s, password: %s", username, password, autoSecret.Username, autoSecret.Password)
	}
}
