ALTER TABLE room_bills
    ADD COLUMN IF NOT EXISTS receipt_image_url TEXT;
