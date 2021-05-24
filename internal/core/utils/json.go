package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// JSONDuplicate check json is duplicate
func JSONDuplicate(d *json.Decoder, path []string) error {
	var duplicate string

	// Get next token from JSON
	t, err := d.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)

	// There's nothing to do for simple values (strings, numbers, bool, nil)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		keys := make(map[string]bool)
		for d.More() {
			// Get field key
			t, err := d.Token()
			if err != nil {
				return err
			}
			key := t.(string)

			// Check for duplicates
			if keys[key] {
				duplicate = duplicate + fmt.Sprint(strings.Join(append(path, "dulicate field: "+key), ","))
			}
			keys[key] = true

			// Check value
			if err := JSONDuplicate(d, append(path, key)); err != nil {
				return err
			}
		}
		// Consume trailing }
		if _, err := d.Token(); err != nil {
			return err
		}

	case '[':
		i := 0
		for d.More() {
			if err := JSONDuplicate(d, append(path, strconv.Itoa(i))); err != nil {
				return err
			}
			i++
		}
		// Consume trailing ]
		if _, err := d.Token(); err != nil {
			return err
		}

	}
	if duplicate != "" {
		return errors.New(duplicate)
	}
	return nil
}

// PrintFormatJSON print format json
func PrintFormatJSON(n interface{}) {
	b, _ := json.MarshalIndent(n, "", "\t")
	os.Stdout.Write(b)
}

// ReadJSONFile read json file
func ReadJSONFile(path string, entities interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), entities)
	if err != nil {
		return err
	}

	return nil
}
