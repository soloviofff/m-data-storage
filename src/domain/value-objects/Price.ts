// Positive price value
export type Price = number & { readonly __brand: 'Price' };

export function price(value: number): Price {
	if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) {
		throw new Error('Price must be a finite positive number');
	}
	return value as Price;
}
