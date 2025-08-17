export interface Broker {
	id: number;
	code: string;
	name: string;
}

export interface Instrument {
	id: number;
	symbol: string;
	name?: string;
}

export interface InstrumentMapping {
	id: number;
	brokerId: number;
	instrumentId: number;
	externalSymbol: string;
}
