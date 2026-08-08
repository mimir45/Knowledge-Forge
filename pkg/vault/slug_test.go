package vault

import "testing"

func TestSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Kafka consumer group rebalancing", "kafka-consumer-group-rebalancing"},
		{"Hibernate @Column mismatch", "hibernate-column-mismatch"},
		{"Keyset / cursor pagination", "keyset-cursor-pagination"},
		{"  leading and trailing  ", "leading-and-trailing"},
		{"Türkçe başlık: şu ğ ı ö ü", "turkce-baslik-su-g-i-o-u"},
		{"C++ vs C#", "cplusplus-vs-csharp"},
		{"Spring Boot 4 + OpenAPI", "spring-boot-4-plus-openapi"},
		{"UPPER CASE TITLE", "upper-case-title"},
		{"multiple---dashes", "multiple-dashes"},
		{"", ""},
		{"!!!", ""},
	}
	for _, c := range cases {
		if got := Slug(c.in); got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestSlugDeterministic is the property the phase spec names explicitly: same input,
// same slug, every time and in every process.
func TestSlugDeterministic(t *testing.T) {
	const title = "Testcontainers Docker socket resolution on colima"
	first := Slug(title)
	for i := 0; i < 1000; i++ {
		if got := Slug(title); got != first {
			t.Fatalf("iteration %d: %q != %q", i, got, first)
		}
	}
}

func TestSlugTruncation(t *testing.T) {
	long := "a very long title that keeps going and going and going well past the eighty " +
		"byte limit the schema imposes on slugs"
	got := Slug(long)
	if len(got) > 80 {
		t.Errorf("len = %d, want <= 80", len(got))
	}
	if !IsSlug(got) {
		t.Errorf("truncated slug %q is not schema-valid", got)
	}
}

func TestSlugUnique(t *testing.T) {
	taken := map[string]bool{"soft-delete": true, "soft-delete-2": true}
	cases := []struct{ in, want string }{
		{"Soft delete", "soft-delete-3"},
		{"Soft deletion", "soft-deletion"},
		{"!!!", "untitled"},
	}
	for _, c := range cases {
		if got := SlugUnique(c.in, taken); got != c.want {
			t.Errorf("SlugUnique(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsSlug(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"kafka-consumer-group", true},
		{"a", true},
		{"a1-b2", true},
		{"Kafka", false},
		{"-leading", false},
		{"trailing-", false},
		{"double--dash", false},
		{"under_score", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsSlug(c.in); got != c.want {
			t.Errorf("IsSlug(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
