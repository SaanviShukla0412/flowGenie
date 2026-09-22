package database

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/SaanviShukla0412/flowGenie/internal/execution"
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

func CreateExecution(
	conn *pgx.Conn,
	id string,
	workflowID string,
	status string,
) error {
	_, err := conn.Exec(
		context.Background(),
		`INSERT INTO executions (id, workflow_id, status)
		 VALUES ($1, $2, $3)`,
		id,
		workflowID,
		status,
	)
	return err
}

func UpdateExecutionStatus(
	conn *pgx.Conn,
	id string,
	status string,
) error {
	_, err := conn.Exec(
		context.Background(),
		`UPDATE executions
		 SET status = $1,
		     updated_at = NOW()
		 WHERE id = $2`,
		status,
		id,
	)
	return err
}

func CreateExecutionStep(
	conn *pgx.Conn,
	id string,
	executionID string,
	stepName string,
	stepType string,
	status string,
) error {
	_, err := conn.Exec(
		context.Background(),
		`INSERT INTO execution_steps (
			id,
			execution_id,
			step_name,
			step_type,
			status
		)
		VALUES ($1, $2, $3, $4, $5)`,
		id,
		executionID,
		stepName,
		stepType,
		status,
	)

	return err
}

func UpdateExecutionStep(
	conn *pgx.Conn,
	id string,
	status string,
	output *string,
	stepError *string,
) error {
	_, err := conn.Exec(
		context.Background(),
		`UPDATE execution_steps
		 SET status = $1,
		     output = $2,
		     error = $3,
		     completed_at = NOW()
		 WHERE id = $4`,
		status,
		output,
		stepError,
		id,
	)

	return err
}

func GetExecutionStepsByExecutionID(
	conn *pgx.Conn,
	executionID string,
) ([]execution.ExecutionStep, error) {
	rows, err := conn.Query(
		context.Background(),
		`SELECT
			id,
			execution_id,
			step_name,
			step_type,
			status,
			output,
			error,
			started_at,
			completed_at
		 FROM execution_steps
		 WHERE execution_id = $1
		 ORDER BY started_at ASC`,
		executionID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	steps := make([]execution.ExecutionStep, 0)

	for rows.Next() {
		var step execution.ExecutionStep

		err := rows.Scan(
			&step.ID,
			&step.ExecutionID,
			&step.StepName,
			&step.StepType,
			&step.Status,
			&step.Output,
			&step.Error,
			&step.StartedAt,
			&step.CompletedAt,
		)

		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func GetExecutionByID(
	conn *pgx.Conn,
	id string,
) (*execution.Execution, error) {
	var exec execution.Execution

	err := conn.QueryRow(
		context.Background(),
		`SELECT id, workflow_id, status, created_at, updated_at
		 FROM executions
		 WHERE id = $1`,
		id,
	).Scan(
		&exec.ID,
		&exec.WorkflowID,
		&exec.Status,
		&exec.CreatedAt,
		&exec.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &exec, nil
}

func GetExecutionsByWorkflowID(
	conn *pgx.Conn,
	workflowID string,
) ([]execution.Execution, error) {
	rows, err := conn.Query(
		context.Background(),
		`SELECT id, workflow_id, status, created_at, updated_at
		 FROM executions
		 WHERE workflow_id = $1
		 ORDER BY created_at DESC`,
		workflowID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executions := make([]execution.Execution, 0)

	for rows.Next() {
		var exec execution.Execution

		err := rows.Scan(
			&exec.ID,
			&exec.WorkflowID,
			&exec.Status,
			&exec.CreatedAt,
			&exec.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		executions = append(executions, exec)
	}

	return executions, nil
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
