package gocan

import (
	"testing"
	"unsafe"
)

func TestMiniCANFDPlatformConfigLayouts(t *testing.T) {
	if got := unsafe.Sizeof(miniCANFDLinuxConfig{}); got != 20 {
		t.Fatalf("Linux config size = %d, want 20", got)
	}
	if got := unsafe.Sizeof(miniCANFDWindowsConfig{}); got != 20 {
		t.Fatalf("Windows config size = %d, want 20", got)
	}

	var windowsConfig miniCANFDWindowsConfig
	for name, gotWant := range map[string]struct{ got, want uintptr }{
		"Config":   {unsafe.Offsetof(windowsConfig.Config), 16},
		"Model":    {unsafe.Offsetof(windowsConfig.Model), 17},
		"Cantype":  {unsafe.Offsetof(windowsConfig.Cantype), 18},
		"Reserved": {unsafe.Offsetof(windowsConfig.Reserved), 19},
	} {
		if gotWant.got != gotWant.want {
			t.Errorf("Windows config %s offset = %d, want %d", name, gotWant.got, gotWant.want)
		}
	}
}

func TestMiniCANFDWindowsConfigConversion(t *testing.T) {
	logical := &miniCANFDConfig{
		NomBaud: 1_000_000,
		DatBaud: 5_000_000,
		NomPre:  0x1234,
		Config:  0x06,
		Model:   0,
		Cantype: 1,
	}
	converted := miniCANFDWindowsConfigFrom(logical)
	if converted.NomBaud != logical.NomBaud || converted.DatBaud != logical.DatBaud {
		t.Fatalf("bitrate conversion = %#v", converted)
	}
	if converted.NomPre != 0x34 {
		t.Fatalf("Windows NomPre = %#x, want low byte 0x34", converted.NomPre)
	}
	if converted.Config != 0x06 || converted.Model != 0 || converted.Cantype != 1 || converted.Reserved != 0 {
		t.Fatalf("Windows control fields = Config %#x Model %#x Cantype %#x Reserved %#x", converted.Config, converted.Model, converted.Cantype, converted.Reserved)
	}
}

func TestMiniCANFDLinuxConfigConversionPreservesNomPre(t *testing.T) {
	logical := &miniCANFDConfig{NomBaud: 1_000_000, DatBaud: 5_000_000, NomPre: 0x1234, Config: 0x07, Cantype: 1}
	converted := miniCANFDLinuxConfigFrom(logical)
	if converted.NomPre != logical.NomPre || converted.Config != logical.Config || converted.Cantype != logical.Cantype {
		t.Fatalf("Linux conversion = %#v", converted)
	}
}
