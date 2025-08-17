import { UnixMinute, unixMinuteFromEpochSeconds } from '../value-objects/UnixMinute';

export interface TimeService {
	nowUnixMinute(): UnixMinute;
}

export class SystemTimeService implements TimeService {
	nowUnixMinute(): UnixMinute {
		const nowSeconds = Math.floor(Date.now() / 1000);
		const aligned = nowSeconds - (nowSeconds % 60);
		return unixMinuteFromEpochSeconds(aligned);
	}
}
