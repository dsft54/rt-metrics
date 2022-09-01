// Package cryptokey implents rsa encryption/decryption of byte slice by ssh keypair

package cryptokey

import (
	"crypto/rsa"
	"math/big"
	"reflect"
	"testing"
)

func TestParsePublicKey(t *testing.T) {
	rsaNum := new(big.Int)
	rsaNum, _ = rsaNum.SetString("3808756663521708221493687987310170449524502114328808616680050238211106236130525240379961032346626750876892976484620189878917911272178522246012264215274827324277165214412139270306065352571013830775875942425019221688683050721952608315887907380545179405567179473508073076175863629748242630236953479392713417377097664923836777815946622537107715814086356094988199726348183006498113514100780469344393965667754407253289540098802716869644037454535761240796183331864805400539161542880202060318485286998694549117013690800594244038899391756680840753030305346450570793551931029647055105956109036237550880943421280783848898595061326133945698457731533688848471012158440124831746275591855869539856894222102677938852155575164434598168395255937885723387342986403129113000548695748877813923121208911851622889997926617754866348321424537579268271131223774747219412029290773126385037307585685550066753641286442607136576971788958804031198831662141", 10)
	tests := []struct {
		name    string
		path    string
		want    *rsa.PublicKey
		wantErr bool
	}{
		{
			name: "normal",
			path: "test.pub",
			want: &rsa.PublicKey{
				N: rsaNum,
				E: 65537,
			},
			wantErr: false,
		},
		{
			name:    "open file failed",
			path:    "pub.test",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ssh authorized key parse failed",
			path:    "test_f.pub",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePublicKey(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePrivateKey(t *testing.T) {
	rsaNum := new(big.Int)
	rsaNum, _ = rsaNum.SetString("3808756663521708221493687987310170449524502114328808616680050238211106236130525240379961032346626750876892976484620189878917911272178522246012264215274827324277165214412139270306065352571013830775875942425019221688683050721952608315887907380545179405567179473508073076175863629748242630236953479392713417377097664923836777815946622537107715814086356094988199726348183006498113514100780469344393965667754407253289540098802716869644037454535761240796183331864805400539161542880202060318485286998694549117013690800594244038899391756680840753030305346450570793551931029647055105956109036237550880943421280783848898595061326133945698457731533688848471012158440124831746275591855869539856894222102677938852155575164434598168395255937885723387342986403129113000548695748877813923121208911851622889997926617754866348321424537579268271131223774747219412029290773126385037307585685550066753641286442607136576971788958804031198831662141", 10)
	tests := []struct {
		name    string
		path    string
		want    *rsa.PrivateKey
		wantErr bool
	}{
		{
			name: "normal",
			path: "test",
			want: &rsa.PrivateKey{
				PublicKey: rsa.PublicKey{
					N: rsaNum,
					E: 65537,
				},
			},
			wantErr: false,
		},
		{
			name:    "open file failed",
			path:    "tset",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "private key parse failed",
			path:    "test_f",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrivateKey(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				if !reflect.DeepEqual(got.PublicKey.N, tt.want.PublicKey.N) {
					t.Errorf("ParsePrivateKey() = %v, want %v", got.PublicKey, tt.want)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParsePrivateKey() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEncryptMessage(t *testing.T) {
	pub, err := ParsePublicKey("test.pub")
	if err != nil {
		t.Log("Normal key parsed")
	}
	pubf, err := ParsePublicKey("test_f.pub")
	if err != nil {
		t.Log("Damaged key parsed")
	}
	tests := []struct {
		name    string
		data    []byte
		pub     *rsa.PublicKey
		want    int
		wantErr bool
	}{
		{
			name:    "normal",
			data:    []byte("message"),
			pub:     pub,
			want:    384,
			wantErr: false,
		},
		{
			name:    "damaged",
			data:    []byte("message"),
			pub:     pubf,
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptMessage(tt.data, tt.pub)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), tt.want) {
				t.Errorf("EncryptMessage() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestDecryptMessage(t *testing.T) {
	pub, err := ParsePublicKey("test.pub")
	if err != nil {
		t.Log("Normal key parsed")
	}
	priv, err := ParsePrivateKey("test")
	if err != nil {
		t.Log("Normal key parsed")
	}
	privf, err := ParsePrivateKey("test_f")
	if err != nil {
		t.Log("Damaged key parsed")
	}
	enc, err := EncryptMessage([]byte("message"), pub)
	if err != nil {
		t.Error("Failed to encrypt data for tests")
	}
	tests := []struct {
		name    string
		data    []byte
		private *rsa.PrivateKey
		step    int
		want    []byte
		wantErr bool
	}{
		{
			name: "normal",
			data: enc,
			private: priv,
			step: 384,
			want: []byte("message"),
			wantErr: false,
		},
		{
			name: "damaged keys",
			data: enc,
			private: privf,
			step: 384,
			want: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptMessage(tt.data, tt.private, tt.step)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
