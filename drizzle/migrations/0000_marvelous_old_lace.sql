CREATE SCHEMA IF NOT EXISTS "registry";
--> statement-breakpoint
CREATE SCHEMA IF NOT EXISTS "timeseries";
--> statement-breakpoint
CREATE TABLE "registry"."audits" (
	"id" serial PRIMARY KEY NOT NULL,
	"ts" timestamp with time zone DEFAULT now() NOT NULL,
	"actor" text,
	"action" text,
	"details" text
);
--> statement-breakpoint
CREATE TABLE "registry"."brokers" (
	"id" serial PRIMARY KEY NOT NULL,
	"code" text NOT NULL,
	"name" text NOT NULL,
	CONSTRAINT "brokers_code_unique" UNIQUE("code")
);
--> statement-breakpoint
CREATE TABLE "registry"."instrument_mappings" (
	"id" serial PRIMARY KEY NOT NULL,
	"broker_id" integer NOT NULL,
	"instrument_id" integer NOT NULL,
	"external_symbol" text NOT NULL
);
--> statement-breakpoint
CREATE TABLE "registry"."instruments" (
	"id" serial PRIMARY KEY NOT NULL,
	"symbol" text NOT NULL,
	"name" text,
	CONSTRAINT "instruments_symbol_unique" UNIQUE("symbol")
);
--> statement-breakpoint
CREATE TABLE "timeseries"."ohlcv" (
	"broker_id" integer NOT NULL,
	"instrument_id" integer NOT NULL,
	"ts" timestamp with time zone NOT NULL,
	"o" real,
	"h" real,
	"l" real,
	"c" real,
	"v" double precision
);
--> statement-breakpoint
CREATE TABLE "registry"."tasks" (
	"id" uuid PRIMARY KEY NOT NULL,
	"broker_id" integer NOT NULL,
	"instrument_id" integer NOT NULL,
	"from_ts" timestamp with time zone NOT NULL,
	"to_ts" timestamp with time zone NOT NULL,
	"status" text NOT NULL,
	"priority" integer DEFAULT 0 NOT NULL,
	"idempotency_key" text NOT NULL,
	"reserved_until" timestamp with time zone,
	CONSTRAINT "tasks_idempotency_key_unique" UNIQUE("idempotency_key")
);
--> statement-breakpoint
ALTER TABLE "registry"."instrument_mappings" ADD CONSTRAINT "instrument_mappings_broker_id_brokers_id_fk" FOREIGN KEY ("broker_id") REFERENCES "registry"."brokers"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "registry"."instrument_mappings" ADD CONSTRAINT "instrument_mappings_instrument_id_instruments_id_fk" FOREIGN KEY ("instrument_id") REFERENCES "registry"."instruments"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "registry"."tasks" ADD CONSTRAINT "tasks_broker_id_brokers_id_fk" FOREIGN KEY ("broker_id") REFERENCES "registry"."brokers"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "registry"."tasks" ADD CONSTRAINT "tasks_instrument_id_instruments_id_fk" FOREIGN KEY ("instrument_id") REFERENCES "registry"."instruments"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
CREATE UNIQUE INDEX "instrument_mappings_broker_external_symbol_uniq" ON "registry"."instrument_mappings" USING btree ("broker_id","external_symbol");--> statement-breakpoint
CREATE INDEX "ohlcv_instrument_ts_idx" ON "timeseries"."ohlcv" USING btree ("instrument_id","ts");--> statement-breakpoint
CREATE UNIQUE INDEX "ohlcv_broker_instrument_ts_uniq" ON "timeseries"."ohlcv" USING btree ("broker_id","instrument_id","ts");--> statement-breakpoint
CREATE INDEX "tasks_instrument_id_idx" ON "registry"."tasks" USING btree ("instrument_id");