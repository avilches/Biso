package ops

import "testing"

func TestSlug(t *testing.T) {
	for _, c := range []struct{ name, want string }{
		{"Kex", "kex"},
		{"Mi Proyecto", "mi-proyecto"},
		{"Peña 2026", "pena-2026"},
		{"My project", "my-project"},
		{"  spaced  out  ", "spaced-out"},
		{"a/b\\c", "a-b-c"},
		{"2026", "2026"},
		{"///", ""},
		{"Café", "cafe"},
		{"mi-proyecto-2", "mi-proyecto-2"},
	} {
		if got := Slug(c.name); got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestDerivePrefix(t *testing.T) {
	for _, c := range []struct{ name, want string }{
		{"mi-proyecto-2", "MIPROYECTO"},
		{"My project", "MYPROJECT"},
		{"Peña", "PENA"},
		{"Café", "CAFE"},
	} {
		got, err := DerivePrefix(c.name)
		if err != nil {
			t.Fatalf("DerivePrefix(%q) failed: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("DerivePrefix(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestANameWithNoLetterHasNoPrefixToDerive(t *testing.T) {
	for _, name := range []string{"2026", "///", ""} {
		if _, err := DerivePrefix(name); err == nil {
			t.Errorf("DerivePrefix(%q) invented a prefix", name)
		} else if err.ExitCode != 2 || err.Code != "invalid_prefix" {
			t.Errorf("DerivePrefix(%q) = %d/%s, want 2/invalid_prefix", name, err.ExitCode, err.Code)
		}
	}
}

func TestValidatePrefixTakesLettersOnly(t *testing.T) {
	for _, prefix := range []string{"MYP", "myp", "Myp"} {
		if err := ValidatePrefix(prefix); err != nil {
			t.Errorf("ValidatePrefix(%q) rejected it: %v", prefix, err)
		}
	}
	for _, prefix := range []string{"MY-P", "MYP1", "", "MY P", "MYÑ"} {
		if err := ValidatePrefix(prefix); err == nil {
			t.Errorf("ValidatePrefix(%q) accepted it", prefix)
		}
	}
}

func TestTheSlugAndThePrefixFailForDifferentReasons(t *testing.T) {
	// "2026" gives a valid slug and no prefix at all, and "///" gives
	// neither, which is why both are checked
	// (docs/spec/resolucion-del-tablero.md#cómo-se-deriva-el-nombre-de-la-carpeta).
	if err := ValidateSlug("2026"); err != nil {
		t.Errorf("the slug of \"2026\" is valid: %v", err)
	}
	if _, err := DerivePrefix("2026"); err == nil {
		t.Error("\"2026\" has no prefix to derive")
	}
	if err := ValidateSlug("///"); err == nil {
		t.Error("\"///\" leaves nothing to name a folder with")
	}
}

func TestFolderName(t *testing.T) {
	if got := FolderName("Mi Proyecto", "3f9a2b1c"); got != "mi-proyecto-3f9a2b1c" {
		t.Errorf("FolderName = %q", got)
	}
}

func TestRandomIDIsEightLowercaseHexadecimalCharacters(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		id, err := RandomID()
		if err != nil {
			t.Fatal(err)
		}
		if len(id) != 8 {
			t.Fatalf("RandomID() = %q, want eight characters", id)
		}
		for _, r := range id {
			if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') {
				t.Fatalf("RandomID() = %q, want lowercase hexadecimal", id)
			}
		}
		seen[id] = true
	}
	if len(seen) < 40 {
		t.Errorf("fifty identifiers gave only %d different ones", len(seen))
	}
}
