// A nominally-typed identifier for brokers
export type BrokerId = number & { readonly __brand: 'BrokerId' };

export function brokerId(value: number): BrokerId {
	if (!Number.isInteger(value) || value <= 0) {
		throw new Error('BrokerId must be a positive integer');
	}
	return value as BrokerId;
}
