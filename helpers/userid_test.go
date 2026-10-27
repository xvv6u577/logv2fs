package helper

import "testing"

func TestIsUUIDEmailAsId(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"01980932db59713b878f8ab420cdbb94", true},
		{"ABCDEF0123456789abcdef0123456789", true},
		{"f8a23e9a86564c3cb4bfd5a099172888", true},
		{"fengzixuan", false},
		{"c39yu3", false},
		{"casterasadmin", false},
		{"1m5y1axc", false},
		{"09u627xri5", false},
		{"", false},
		{"01980932db59713b878f8ab420cdbb9", false},
		{"01980932db59713b878f8ab420cdbb941", false},
		{"01980932-db59-713b-878f-8ab420cdbb94", false},
	}
	for _, tc := range cases {
		if got := IsUUIDEmailAsId(tc.id); got != tc.want {
			t.Fatalf("IsUUIDEmailAsId(%q)=%v want %v", tc.id, got, tc.want)
		}
	}
}

func TestNewUUIDEmailAsId(t *testing.T) {
	id, err := NewUUIDEmailAsId()
	if err != nil {
		t.Fatal(err)
	}
	if !IsUUIDEmailAsId(id) {
		t.Fatalf("generated id not uuid-style: %q", id)
	}
}

func TestUserIDFilterShape(t *testing.T) {
	raw := UserIDFilter("abc")
	if raw["$or"] == nil {
		t.Fatal("missing $or")
	}
}
