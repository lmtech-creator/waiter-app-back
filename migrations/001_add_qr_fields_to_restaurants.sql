ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS qr_banner_text TEXT NOT NULL DEFAULT 'Escaneá y llamá al mozo',
  ADD COLUMN IF NOT EXISTS qr_footer_items JSONB NOT NULL DEFAULT '["Llamar al mozo","Pedir la cuenta","Dejar reseña"]'::jsonb;
