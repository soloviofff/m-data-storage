-- Rename brokers.code -> brokers.system_name and keep uniqueness (idempotent)
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'registry' AND table_name = 'brokers' AND column_name = 'code'
  ) THEN
    ALTER TABLE "registry"."brokers" RENAME COLUMN "code" TO "system_name";
  END IF;
END $$;
-- Drop old unique if present and create new unique index
ALTER TABLE "registry"."brokers" DROP CONSTRAINT IF EXISTS "brokers_code_unique";
ALTER TABLE "registry"."brokers" ADD CONSTRAINT "brokers_system_name_unique" UNIQUE ("system_name");

-- Add instruments.broker_id and constraints
ALTER TABLE "registry"."instruments" ADD COLUMN IF NOT EXISTS "broker_id" integer;
ALTER TABLE "registry"."instruments"
  ADD CONSTRAINT "instruments_broker_id_brokers_id_fk"
  FOREIGN KEY ("broker_id") REFERENCES "registry"."brokers"("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- Replace unique(symbol) with unique(broker_id, symbol)
ALTER TABLE "registry"."instruments" DROP CONSTRAINT IF EXISTS "instruments_symbol_unique";
ALTER TABLE "registry"."instruments" ADD CONSTRAINT "instruments_broker_symbol_uniq" UNIQUE ("broker_id","symbol");
CREATE INDEX IF NOT EXISTS "instruments_broker_id_idx" ON "registry"."instruments" ("broker_id");
