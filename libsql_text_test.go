package libsql

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func referenceText(value string) any {
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	} {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return parsed
		}
	}
	return value
}

func textCases(t *testing.T) []string {
	t.Helper()
	file, err := os.Open("testdata/text-values.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var cases []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cases = append(cases, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return cases
}

func TestTextConversion(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("libsql", "file:"+t.TempDir()+"/text.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, value := range textCases(t) {
		t.Run(fmt.Sprintf("given %q, when read, then its type and value match", value), func(t *testing.T) {
			t.Parallel()
			want := referenceText(value)
			var got any
			if err := db.QueryRowContext(context.Background(), "SELECT ?", value).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v (%T), want %#v (%T)", got, got, want, want)
			}
		})
	}
}

func FuzzDatePrefix(f *testing.F) {
	f.Fuzz(func(t *testing.T, value string) {
		want := referenceText(value)
		got := any(value)
		if hasDatePrefix(value) {
			got = referenceText(value)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v (%T), want %#v (%T)", got, got, want, want)
		}
	})
}
