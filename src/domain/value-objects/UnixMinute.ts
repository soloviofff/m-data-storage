// Unix timestamp aligned to minute (seconds since epoch divisible by 60)
export type UnixMinute = number & { readonly __brand: 'UnixMinute' };

export function unixMinuteFromEpochSeconds(epochSeconds: number): UnixMinute {
	if (!Number.isInteger(epochSeconds) || epochSeconds < 0) {
		throw new Error('UnixMinute must be built from a non-negative integer epoch seconds');
	}
	if (epochSeconds % 60 !== 0) {
		throw new Error('UnixMinute must be aligned to a minute (epochSeconds % 60 === 0)');
	}
	return epochSeconds as UnixMinute;
}
