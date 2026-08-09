package store

import "database/sql"

// Rows loads every cached note keyed by rel, in three queries rather than one per note.
// `forge recall` reads the whole cache on every run, and 91 round trips through the
// driver costs more than the file reads the cache was meant to avoid.
//
// Links are not loaded: recall does not need them, and the join is the expensive one.
func (s *Store) Rows() (map[string]Row, error) {
	rows, err := s.notes()
	if err != nil {
		return nil, err
	}
	for _, t := range []string{"note_stack", "note_tags"} {
		if err := s.attach(rows, t); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

const notesSQL = `SELECT rel,slug,title,type,confidence,updated,verified,
 freshness_days,mtime,size,valid FROM notes`

func (s *Store) notes() (map[string]Row, error) {
	rs, err := s.DB.Query(notesSQL)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := map[string]Row{}
	for rs.Next() {
		r, err := scanRow(rs)
		if err != nil {
			return nil, err
		}
		out[r.Rel] = r
	}
	return out, rs.Err()
}

func scanRow(rs *sql.Rows) (Row, error) {
	var r Row
	err := rs.Scan(&r.Rel, &r.Slug, &r.Title, &r.Type, &r.Confidence, &r.Updated,
		&r.Verified, &r.FreshnessDays, &r.MTime, &r.Size, &r.Valid)
	return r, err
}

// attach folds a child table's values onto the rows already loaded.
func (s *Store) attach(rows map[string]Row, table string) error {
	rs, err := s.DB.Query("SELECT rel,value FROM " + table)
	if err != nil {
		return err
	}
	defer rs.Close()
	for rs.Next() {
		var rel, val string
		if err := rs.Scan(&rel, &val); err != nil {
			return err
		}
		appendChild(rows, table, rel, val)
	}
	return rs.Err()
}

func appendChild(rows map[string]Row, table, rel, val string) {
	r, ok := rows[rel]
	if !ok {
		return // orphaned child row; Prune will collect it
	}
	if table == "note_stack" {
		r.Stack = append(r.Stack, val)
	} else {
		r.Tags = append(r.Tags, val)
	}
	rows[rel] = r
}
