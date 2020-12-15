package queries

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

const (
	psqlVarRE = `[^:]:['"]?([A-Za-z][A-Za-z0-9_]*)['"]?`
)

type (
	QueryStore struct {
		queries      map[string]*Query
		OpenFunc     func(string, func(io.Reader) error) error
		WalkImplFunc func(p string, wf filepath.WalkFunc) error
	}

	Query struct {
		Raw          string
		OrdinalQuery string
		Mapping      map[string]int
	}
)

// NewQueryStore setups new query store
func NewQueryStore() *QueryStore {
	return &QueryStore{
		queries:      make(map[string]*Query),
		WalkImplFunc: filepath.Walk,
		OpenFunc: func(file string, load func(io.Reader) error) error {
			f, err := os.Open(file)
			if err != nil {
				return err
			}

			defer f.Close()

			return load(f)
		},
	}
}

// LoadFromFile loads query/queries from specified file
func (s *QueryStore) LoadFromFile(file string) (err error) {
	return s.OpenFunc(file, s.loadQueriesFromFile)
}

// LoadFromGlob loads queries from all .sql files in root
func (s *QueryStore) LoadFromWalk(root string) (err error) {
	var (
		matches []string
	)

	matches = make([]string, 0)

	err = s.WalkImplFunc(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".sql" {
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	for _, match := range matches {
		loadErr := s.LoadFromFile(match)
		if loadErr != nil {
			return loadErr
		}
	}

	return nil
}

// MustHaveQuery returns query or panics on error
func (s *QueryStore) MustHaveQuery(name string) *Query {
	query, err := s.Query(name)
	if err != nil {
		panic(err)
	}

	return query
}

// Query retrieve query by given name
func (s *QueryStore) Query(name string) (*Query, error) {
	query, ok := s.queries[name]
	if !ok {
		return nil, fmt.Errorf("Query '%s' not found", name)
	}

	return query, nil
}

func (s *QueryStore) loadQueriesFromFile(r io.Reader) error {
	scanner := &Scanner{}
	newQueries := scanner.Run(bufio.NewScanner(r))

	for name, query := range newQueries {
		// insert query (but check whatever it already exists)
		if _, ok := s.queries[name]; ok {
			return fmt.Errorf("Query '%s' already exists", name)
		}

		q := NewQuery(query)

		s.queries[name] = q
	}

	return nil
}

func NewQuery(query string) *Query {
	var (
		position int = 1
	)

	q := Query{
		Raw: query,
	}

	mapping := make(map[string]int)

	r, _ := regexp.Compile(psqlVarRE)
	matches := r.FindAllStringSubmatch(query, -1)

	for _, match := range matches {
		variable := match[1]

		if _, ok := mapping[variable]; !ok {
			mapping[variable] = position
			position++
		}
	}

	// replace the variable with ordinal markers
	for name, ord := range mapping {
		r, _ := regexp.Compile(fmt.Sprintf(`:["']?%s["']?`, name))
		query = r.ReplaceAllLiteralString(query, fmt.Sprintf("$%d", ord))
	}

	q.OrdinalQuery = query
	q.Mapping = mapping

	return &q
}

// Query returns ordinal query
func (q *Query) Query() string {
	return q.OrdinalQuery
}

// Prepare the arguments for the ordinal query. Missing arguments will
// be returned as nil
func (q *Query) Prepare(args map[string]interface{}) []interface{} {
	type kv struct {
		Name string
		Ord  int
	}

	var components []interface{}

	// number of components is query and ordinal mapping count
	components = make([]interface{}, len(q.Mapping))
	var params []kv
	for k, v := range q.Mapping {
		params = append(params, kv{k, v})
	}

	sort.Slice(params, func(i, j int) bool {
		return params[i].Ord < params[j].Ord
	})

	for i, param := range params {
		components[i] = args[param.Name]
	}

	return components
}
