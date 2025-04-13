CREATE TRIGGER trg_edited_at_users
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_edited_at();

CREATE TRIGGER trg_edited_at_tutor_profiles
    BEFORE UPDATE ON tutor_profiles
    FOR EACH ROW
    EXECUTE FUNCTION set_edited_at();

CREATE TRIGGER trg_edited_at_tutor_students
    BEFORE UPDATE ON tutor_students
    FOR EACH ROW
    EXECUTE FUNCTION set_edited_at();
