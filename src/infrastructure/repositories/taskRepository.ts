import { getPool } from '../db/client';
import { randomUUID } from 'crypto';

export type TaskStatus = 'queued' | 'reserved' | 'done' | 'failed';

export interface TaskRow {
	id: string;
	broker_id: number;
	instrument_id: number;
	from_ts: string;
	to_ts: string;
	status: TaskStatus;
	priority: number;
	idempotency_key: string;
	reserved_until: string | null;
}

export interface ReserveParams {
	brokerId?: number;
	instrumentIds?: number[];
	limit?: number;
	leaseSeconds?: number;
}

export async function reserveNextTasks(params: ReserveParams): Promise<TaskRow[]> {
	const { brokerId, instrumentIds, limit = 10, leaseSeconds = 60 } = params;
	const pool = getPool();
	const where: string[] = [
		`status = 'queued'`,
		`(reserved_until IS NULL OR reserved_until < now())`,
	];
	const values: Array<number | string | Date | number[]> = [];
	let idx = 1;

	if (typeof brokerId === 'number') {
		where.push(`broker_id = $${idx++}`);
		values.push(brokerId);
	}
	if (instrumentIds && instrumentIds.length > 0) {
		where.push(`instrument_id = ANY($${idx++}::int[])`);
		values.push(instrumentIds);
	}
	const whereSql = where.join(' AND ');

	// CTE to select candidates with SKIP LOCKED, then update to reserved with lease
	const sql = `
    WITH candidates AS (
      SELECT id
      FROM registry.tasks
      WHERE ${whereSql}
      ORDER BY priority DESC, from_ts ASC
      LIMIT ${limit}
      FOR UPDATE SKIP LOCKED
    )
    UPDATE registry.tasks t
    SET status = 'reserved', reserved_until = now() + ($${idx} || ' seconds')::interval
    FROM candidates c
    WHERE t.id = c.id
    RETURNING t.*
  `;
	values.push(leaseSeconds);
	const res = await pool.query(sql, values);
	return (res.rows ?? []) as TaskRow[];
}

export async function completeTask(
	id: string,
	status: Exclude<TaskStatus, 'queued' | 'reserved'>,
): Promise<TaskRow | null> {
	const pool = getPool();
	const sql = `
    UPDATE registry.tasks
    SET status = $2, reserved_until = NULL
    WHERE id = $1 AND status = 'reserved'
    RETURNING *
  `;
	const res = await pool.query(sql, [id, status]);
	if ((res.rowCount ?? 0) === 0) return null;
	return res.rows[0] as TaskRow;
}

export interface CreateTaskInput {
	brokerId: number;
	instrumentId: number;
	fromTs: Date;
	toTs: Date;
	priority?: number;
	idempotencyKey: string;
}

export async function createTask(input: CreateTaskInput): Promise<TaskRow> {
	const pool = getPool();
	const sql = `
    INSERT INTO registry.tasks (id, broker_id, instrument_id, from_ts, to_ts, status, priority, idempotency_key)
    VALUES ($1, $2, $3, $4, $5, 'queued', COALESCE($6, 0), $7)
    ON CONFLICT (idempotency_key) DO UPDATE SET
      priority = EXCLUDED.priority
    RETURNING *
  `;
	const id = randomUUID();
	const values = [
		id,
		input.brokerId,
		input.instrumentId,
		input.fromTs,
		input.toTs,
		input.priority ?? 0,
		input.idempotencyKey,
	];
	const res = await pool.query(sql, values);
	return res.rows[0] as TaskRow;
}
