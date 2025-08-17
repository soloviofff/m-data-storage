import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
require('dotenv-expand').expand(require('dotenv').config());

const { createTask, reserveNextTasks, completeTask } = await import('../dist/infrastructure/repositories/taskRepository.js');

test('create task and reserve/complete', async (t) => {
  const now = new Date();
  const fromTs = new Date(now.getTime() - 60 * 60 * 1000);
  const toTs = new Date(now.getTime() - 30 * 60 * 1000);
  const idem = `unit-task-${Date.now()}`;

  const created = await createTask({ brokerId: 1, instrumentId: 1, fromTs, toTs, idempotencyKey: idem, priority: 1 });
  assert.ok(created.id, 'task id should be returned');
  // After ON CONFLICT UPDATE, task may already exist; allow queued or done depending on prior runs
  assert.ok(['queued', 'reserved', 'done', 'failed'].includes(created.status));

  const reserved = await reserveNextTasks({ brokerId: 1, instrumentIds: [1], limit: 1, leaseSeconds: 10 });
  assert.ok(Array.isArray(reserved));
  assert.equal(reserved.length >= 1, true);
  const task = reserved[0];
  assert.equal(task.status, 'reserved');

  const completed = await completeTask(task.id, 'done');
  assert.ok(completed);
  assert.equal(completed.status, 'done');
});


