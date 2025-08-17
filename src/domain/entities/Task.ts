import type { BrokerId } from '../value-objects/BrokerId';
import type { InstrumentId } from '../value-objects/InstrumentId';
import type { UnixMinute } from '../value-objects/UnixMinute';

export type TaskStatus = 'queued' | 'reserved' | 'done' | 'failed';

export interface Task {
	id: string; // uuid
	brokerId: BrokerId;
	instrumentId: InstrumentId;
	fromTs: UnixMinute;
	toTs: UnixMinute;
	priority: number;
	status: TaskStatus;
	idempotencyKey: string;
	reservedUntil?: Date | null;
}
