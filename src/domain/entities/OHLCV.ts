import type { BrokerId } from '../value-objects/BrokerId';
import type { InstrumentId } from '../value-objects/InstrumentId';
import type { UnixMinute } from '../value-objects/UnixMinute';
import type { Price } from '../value-objects/Price';
import type { Volume } from '../value-objects/Volume';

export interface OHLCVBar {
	brokerId: BrokerId;
	instrumentId: InstrumentId;
	ts: UnixMinute;
	o: Price;
	h: Price;
	l: Price;
	c: Price;
	v: Volume;
}
