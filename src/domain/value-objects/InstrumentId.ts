// A nominally-typed identifier for instruments
export type InstrumentId = number & { readonly __brand: 'InstrumentId' };

export function instrumentId(value: number): InstrumentId {
	if (!Number.isInteger(value) || value <= 0) {
		throw new Error('InstrumentId must be a positive integer');
	}
	return value as InstrumentId;
}
