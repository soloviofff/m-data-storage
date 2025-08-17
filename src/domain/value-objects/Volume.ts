// Non-negative volume value
export type Volume = number & { readonly __brand: 'Volume' };

export function volume(value: number): Volume {
	if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
		throw new Error('Volume must be a finite non-negative number');
	}
	return value as Volume;
}
