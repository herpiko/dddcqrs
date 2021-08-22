package main

import (
	"database/sql"
)

type project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (p *project) getProject(db *sql.DB) error {
	return db.QueryRow("SELECT name FROM projects WHERE id=$1",
		p.ID).Scan(&p.Name)
}

func (p *project) updateProject(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE projects SET name=$1 WHERE id=$3",
			p.Name, p.ID)

	return err
}

func (p *project) deleteProject(db *sql.DB) error {
	_, err := db.Exec("UPDATE projects SET deleted_at=NOW() WHERE id=$1", p.ID)
	return err
}

func (p *project) createProject(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO projects(name) VALUES($1) RETURNING id",
		p.Name).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func getProjects(db *sql.DB, start, count int) ([]project, error) {
	rows, err := db.Query(
		"SELECT id, name FROM projects LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	projects := []project{}

	for rows.Next() {
		var p project
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}
