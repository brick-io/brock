package sdksql

import "strings"

//nolint:gochecknoglobals
var Tool tool

type tool struct{}

const (
	_SELECT   = "SELECT"
	_INSERT   = "INSERT"
	_UPDATE   = "UPDATE"
	_DELETE   = "DELETE"
	_CREATE   = "CREATE"
	_ALTER    = "ALTER"
	_DROP     = "DROP"
	_USE      = "USE"
	_ADD      = "ADD"
	_EXEC     = "EXEC"
	_TRUNCATE = "TRUNCATE"
)

// RemoveComment from sql command.
func (tool) RemoveComment(query string) string {
	commentStartIdx, replaces := -1, []string{}

	for i := range query {
		// we found sql comment
		if i-1 >= 0 && i+1 < len(query) && query[i] == '-' && query[i+1] == '-' && query[i-1] != '-' {
			commentStartIdx = i

			continue
		}

		if commentStartIdx > -1 && query[i] == '\n' {
			replaces = append(replaces, query[commentStartIdx:i])
		}
	}

	for _, v := range replaces {
		query = strings.Replace(query, v, "", 1)
	}

	return strings.TrimSpace(query)
}

// IsMultipleCommand is a naive implementation of checking multiple sql command.
func (x tool) IsMultipleCommand(query string) bool {
	validCount := 0

	for _, query := range strings.Split(query, ";") {
		query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
		if x.IsValidCommand(query) {
			validCount++
		}
	}

	return validCount > 1
}

// IsSELECTCommand only valid if starts with SELECT.
func (x tool) IsSELECTCommand(query string) bool {
	var ok bool

	query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
	for _, s := range []string{_SELECT} {
		ok = ok || strings.HasPrefix(query, s) || strings.Contains(query, s)
	}

	return ok
}

// IsDMLCommand only valid if starts with INSERT, UPDATE, DELETE.
func (x tool) IsDMLCommand(query string) bool {
	var ok bool

	query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
	for _, s := range []string{_INSERT, _UPDATE, _DELETE} {
		ok = ok || strings.HasPrefix(query, s)
	}

	return ok
}

// IsDDLCommand only valid if starts with CREATE, ALTER, DROP, USE, ADD, EXEC, TRUNCATE.
func (x tool) IsDDLCommand(query string) bool {
	var ok bool

	query = strings.ToUpper(strings.TrimSpace(x.RemoveComment(query)))
	for _, s := range []string{_CREATE, _ALTER, _DROP, _USE, _ADD, _EXEC, _TRUNCATE} {
		ok = ok || strings.HasPrefix(query, s)
	}

	return ok
}

func (x tool) IsValidCommand(query string) bool {
	return x.IsSELECTCommand(query) || x.IsDMLCommand(query) || x.IsDDLCommand(query)
}

func (x tool) EscapeQuery(query string) string {
	return strings.NewReplacer(
		"(", "\\(",
		")", "\\)",
	).Replace(query)
}
