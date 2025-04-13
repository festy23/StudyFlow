CREATE OR REPLACE FUNCTION set_edited_at()
RETURNS TRIGGER AS $$
BEGIN
	NEW.edited_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;