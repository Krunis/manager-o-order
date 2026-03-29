package sagacoordinator

import (
	"context"
)

func (rep *PostgresSagaRepository) Save(ctx context.Context, state *SagaState) (string, error){
	var id string

	err := rep.pool.QueryRow(ctx, `INSERT INTO saga_states(order_id, status, current_step, payload, error)
								  VALUES($1, $2, $3, $4, $5, $6)
								  RETURNING id`,
					state.OrderID, state.Status, state.CurrentStep, state.Payload, state.Error).Scan(&id)
	if err != nil{
		return "", err
	}

	return id, nil
}

func (rep *PostgresSagaRepository) Update(ctx context.Context, state *SagaState) error{
	_, err := rep.pool.Exec(ctx, `UPDATE saga_states
								  SET status=$1, current_step=$2, payload=$3, error=$4
								  WHERE id=$5`,
								state.Status, state.CurrentStep, state.Payload, state.Error, state.ID)
	if err != nil{
		return err
	}

	return nil
}

func (rep *PostgresSagaRepository) FindByID(ctx context.Context, id string) (*SagaState, error){
	state := &SagaState{}

	row := rep.pool.QueryRow(ctx, `SELECT * FROM saga_states
								   WHERE id=$1`, id)
					
	err := row.Scan(
		state.ID,
		state.OrderID,
		state.Status, 
		state.CurrentStep,
		state.Payload,
		state.Error,
		state.CreatedAt, 
		state.UpdatedAt,
	)
	if err != nil{
		return nil, err
	}

	return state, nil
}

func (rep *PostgresSagaRepository) Close() {
	rep.pool.Close()
}