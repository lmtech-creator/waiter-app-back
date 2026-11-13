-- Plans module: configurable plans and restaurant.plan_id reference.
-- Safe to run on databases that still have restaurants.plan (text) or are fresh.

CREATE TABLE IF NOT EXISTS plans (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT ''
);

-- Ensure at least the default plan exists.
INSERT INTO plans (id, name, description)
VALUES ('00000000-0000-0000-0000-000000000001', 'free', 'Plan gratuito')
ON CONFLICT (name) DO NOTHING;

-- Backfill plans from any existing restaurant plan names.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'restaurants' AND column_name = 'plan'
  ) THEN
    INSERT INTO plans (id, name, description)
    SELECT gen_random_uuid(), trim(p.plan_name), ''
    FROM (
      SELECT DISTINCT plan AS plan_name
      FROM restaurants
      WHERE plan IS NOT NULL AND trim(plan) <> ''
    ) p
    WHERE NOT EXISTS (
      SELECT 1 FROM plans pl WHERE lower(pl.name) = lower(trim(p.plan_name))
    );
  END IF;
END $$;

-- Add plan_id to restaurants and backfill from legacy plan name.
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS plan_id UUID;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'restaurants' AND column_name = 'plan'
  ) THEN
    UPDATE restaurants r
    SET plan_id = p.id
    FROM plans p
    WHERE r.plan_id IS NULL AND lower(p.name) = lower(trim(r.plan));
  END IF;
END $$;

-- Fallback any remaining rows to the default plan.
UPDATE restaurants
SET plan_id = '00000000-0000-0000-0000-000000000001'
WHERE plan_id IS NULL;

ALTER TABLE restaurants ALTER COLUMN plan_id SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_restaurants_plan'
  ) THEN
    ALTER TABLE restaurants
      ADD CONSTRAINT fk_restaurants_plan
      FOREIGN KEY (plan_id) REFERENCES plans(id);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_restaurants_plan_id ON restaurants(plan_id);

-- Drop legacy column once data is migrated.
ALTER TABLE restaurants DROP COLUMN IF EXISTS plan;
