package database

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/SaanviShukla0412/flowGenie/internal/workflow"
)

func Connect() (*pgx.Conn, error) {
	conn, err := pgx.Connect(
		context.Background(),
		"postgres://saanvi.shukla@localhost:5432/flowgenie",
	)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func CreateWorkflow(
	conn *pgx.Conn,
	id string,
	name string,
	steps []byte,
) error {
	_, err := conn.Exec(
		context.Background(),
		`INSERT INTO workflows (id, name, steps)
		 VALUES ($1, $2, $3)`,
		id,
		name,
		steps,
	)
	return err
}

func GetWorkflows(conn *pgx.Conn) ([]workflow.Workflow, error) {
	rows, err := conn.Query(
		context.Background(),
		`SELECT id, name, steps FROM workflows`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workflows []workflow.Workflow
	for rows.Next() {
		var wf workflow.Workflow
		var stepsJSON []byte
		err := rows.Scan(
			&wf.ID,
			&wf.Name,
			&stepsJSON,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(stepsJSON, &wf.Steps)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, wf)
	}
	return workflows, nil
}

func GetWorkflowByID(
	conn *pgx.Conn,
	id string,
) (*workflow.Workflow, error) {
	var wf workflow.Workflow
	var stepsJSON []byte
	err := conn.QueryRow(
		context.Background(),
		`SELECT id, name, steps
		 FROM workflows
		 WHERE id = $1`,
		id,
	).Scan(
		&wf.ID,
		&wf.Name,
		&stepsJSON,
	)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(stepsJSON, &wf.Steps)
	if err != nil {
		return nil, err
	}
	return &wf, nil
}
